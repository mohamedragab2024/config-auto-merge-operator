package controller

import (
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/mohamedragab2024/config-auto-merge-operator/pkg/handler"
	"github.com/mohamedragab2024/config-auto-merge-operator/pkg/metrics"
)

const (
	watchAnnotation = "config-merger.k8s.io/watch"
	targetConfigMap = "config-merger.k8s.io/target"
)

type ConfigMapController struct {
	clientset        kubernetes.Interface
	informer         cache.SharedIndexInformer
	configMapHandler *handler.ConfigMapHandler
}

func NewConfigMapController(clientset kubernetes.Interface) *ConfigMapController {
	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	informer := factory.Core().V1().ConfigMaps().Informer()

	controller := &ConfigMapController{
		clientset:        clientset,
		informer:         informer,
		configMapHandler: handler.NewConfigMapHandler(clientset),
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.handleAdd,
		UpdateFunc: controller.handleUpdate,
		DeleteFunc: controller.handleDelete,
	})

	return controller
}

func (c *ConfigMapController) Run(stopCh <-chan struct{}) {
	logrus.Info("Starting ConfigMap controller")
	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		logrus.Fatal("Timed out waiting for caches to sync")
	}

	<-stopCh
}

func (c *ConfigMapController) handleAdd(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	if value, exists := configMap.Annotations[watchAnnotation]; exists && value == "true" {
		metrics.WatchedConfigMapsCount.WithLabelValues(configMap.Namespace).Inc()
		c.configMapHandler.HandleConfigMapChange(configMap)
	}
}

func (c *ConfigMapController) handleUpdate(oldObj, newObj interface{}) {
	configMap := newObj.(*corev1.ConfigMap)
	if value, exists := configMap.Annotations[watchAnnotation]; exists && value == "true" {
		c.configMapHandler.HandleConfigMapChange(configMap)
	}
}

func (c *ConfigMapController) handleDelete(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	if value, exists := configMap.Annotations[watchAnnotation]; exists && value == "true" {
		metrics.WatchedConfigMapsCount.WithLabelValues(configMap.Namespace).Dec()
		c.configMapHandler.HandleConfigMapDeletion(configMap)
		metrics.ConfigMapOperations.WithLabelValues("delete", "success").Inc()
	}
}
