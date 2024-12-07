package utils

import (
	corev1 "k8s.io/api/core/v1"
)

func MergeConfigMaps(configMaps []corev1.ConfigMap, targetName string) map[string]string {
	mergedData := make(map[string]string)

	for _, cm := range configMaps {
		// Skip the target configmap and configmaps without the watch annotation
		if cm.Name == targetName {
			continue
		}
		if value, exists := cm.Annotations["config-merger.k8s.io/watch"]; !exists || value != "true" {
			continue
		}

		// Merge data
		for key, value := range cm.Data {
			mergedData[cm.Name+"."+key] = value
		}
	}

	return mergedData
}
