package handler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/yourusername/configmap-operator/pkg/utils"
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
	targetName, exists := configMap.Annotations["config-merger.k8s.io/target"]
	if !exists {
		klog.Warningf("ConfigMap %s/%s has no target configmap specified", configMap.Namespace, configMap.Name)
		return
	}

	// Get all configmaps in the namespace
	cms, err := h.clientset.CoreV1().ConfigMaps(configMap.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Error listing configmaps: %v", err)
		return
	}

	// Filter and merge watched configmaps
	mergedData := utils.MergeConfigMaps(cms.Items, targetName)

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
	} else {
		// Update if exists
		_, err = h.clientset.CoreV1().ConfigMaps(configMap.Namespace).Update(context.TODO(), targetCM, metav1.UpdateOptions{})
	}

	if err != nil {
		klog.Errorf("Error updating target configmap: %v", err)
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
		klog.Errorf("Error listing configmaps: %v", err)
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
		klog.Errorf("Error updating target configmap: %v", err)
	}
}
