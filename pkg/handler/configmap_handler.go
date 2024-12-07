package handler

import (
	"context"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/mohamedragab2024/config-auto-merge-operator/pkg/metrics"
	"github.com/mohamedragab2024/config-auto-merge-operator/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type ConfigMapHandler struct {
	clientset kubernetes.Interface
}

func NewConfigMapHandler(clientset kubernetes.Interface) *ConfigMapHandler {
	return &ConfigMapHandler{
		clientset: clientset,
	}
}

func (h *ConfigMapHandler) HandleConfigMapChange(configMap *corev1.ConfigMap) {
	timer := prometheus.NewTimer(metrics.ConfigMapProcessingLatency.WithLabelValues("change"))
	defer timer.ObserveDuration()

	targetName, exists := configMap.Annotations["config-merger.k8s.io/target"]
	if !exists {
		logrus.Warnf("ConfigMap %s/%s has no target configmap specified", configMap.Namespace, configMap.Name)
		metrics.ConfigMapErrors.WithLabelValues("validation").Inc()
		return
	}

	// Get all configmaps in the namespace
	cms, err := h.clientset.CoreV1().ConfigMaps(configMap.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Error listing configmaps: %v", err)
		metrics.ConfigMapErrors.WithLabelValues("kubernetes_api").Inc()
		return
	}

	// Filter and merge watched configmaps
	mergedData := utils.MergeConfigMaps(cms.Items, targetName)

	// Track merged size
	var totalSize int64
	for _, value := range mergedData {
		totalSize += int64(len(value))
	}
	metrics.MergedConfigMapsSize.WithLabelValues(configMap.Namespace, targetName).Set(float64(totalSize))

	// Create or update target configmap
	targetCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName,
			Namespace: configMap.Namespace,
		},
		Data: mergedData,
	}

	_, err = h.clientset.CoreV1().ConfigMaps(configMap.Namespace).Get(context.TODO(), targetName, metav1.GetOptions{})
	if err != nil {
		// Create if not exists
		_, err = h.clientset.CoreV1().ConfigMaps(configMap.Namespace).Create(context.TODO(), targetCM, metav1.CreateOptions{})
		metrics.ConfigMapOperations.WithLabelValues("create", getLabel(err)).Inc()
	} else {
		// Update if exists
		_, err = h.clientset.CoreV1().ConfigMaps(configMap.Namespace).Update(context.TODO(), targetCM, metav1.UpdateOptions{})
		metrics.ConfigMapOperations.WithLabelValues("update", getLabel(err)).Inc()
	}

	if err != nil {
		logrus.Errorf("Error updating target configmap: %v", err)
		metrics.ConfigMapErrors.WithLabelValues("kubernetes_api").Inc()
	}
}

func (h *ConfigMapHandler) HandleConfigMapDeletion(configMap *corev1.ConfigMap) {
	// Similar to HandleConfigMapChange but handles deletion
	targetName, exists := configMap.Annotations["config-merger.k8s.io/target"]
	if !exists {
		return
	}

	// Recompute merged configmap without the deleted one
	cms, err := h.clientset.CoreV1().ConfigMaps(configMap.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Error listing configmaps: %v", err)
		return
	}

	mergedData := utils.MergeConfigMaps(cms.Items, targetName)

	targetCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName,
			Namespace: configMap.Namespace,
		},
		Data: mergedData,
	}

	_, err = h.clientset.CoreV1().ConfigMaps(configMap.Namespace).Update(context.TODO(), targetCM, metav1.UpdateOptions{})
	if err != nil {
		logrus.Errorf("Error updating target configmap: %v", err)
	}
}

func getLabel(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
