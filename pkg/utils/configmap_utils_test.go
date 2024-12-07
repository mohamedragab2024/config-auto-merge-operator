package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeConfigMaps(t *testing.T) {
	tests := []struct {
		name       string
		configMaps []corev1.ConfigMap
		targetName string
		want       map[string]string
	}{
		{
			name: "merge two configmaps",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config1",
						Annotations: map[string]string{
							"config-merger.k8s.io/watch": "true",
						},
					},
					Data: map[string]string{
						"key1": "value1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config2",
						Annotations: map[string]string{
							"config-merger.k8s.io/watch": "true",
						},
					},
					Data: map[string]string{
						"key2": "value2",
					},
				},
			},
			targetName: "merged-config",
			want: map[string]string{
				"config1.key1": "value1",
				"config2.key2": "value2",
			},
		},
		{
			name: "skip configmap without watch annotation",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config1",
						Annotations: map[string]string{
							"config-merger.k8s.io/watch": "true",
						},
					},
					Data: map[string]string{
						"key1": "value1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config2",
					},
					Data: map[string]string{
						"key2": "value2",
					},
				},
			},
			targetName: "merged-config",
			want: map[string]string{
				"config1.key1": "value1",
			},
		},
		{
			name: "skip target configmap",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "merged-config",
						Annotations: map[string]string{
							"config-merger.k8s.io/watch": "true",
						},
					},
					Data: map[string]string{
						"key1": "value1",
					},
				},
			},
			targetName: "merged-config",
			want:       map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeConfigMaps(tt.configMaps, tt.targetName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeConfigMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}
