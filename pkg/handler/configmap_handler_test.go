package handler

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHandleConfigMapChange(t *testing.T) {
	tests := []struct {
		name           string
		existingCMs    []corev1.ConfigMap
		inputCM        *corev1.ConfigMap
		expectedMerged map[string]string
		expectCreation bool
	}{
		{
			name: "create new merged configmap",
			existingCMs: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "config1",
						Namespace: "default",
						Annotations: map[string]string{
							"config-merger.k8s.io/watch":  "true",
							"config-merger.k8s.io/target": "merged-config",
						},
					},
					Data: map[string]string{
						"key1": "value1",
					},
				},
			},
			inputCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "config2",
					Namespace: "default",
					Annotations: map[string]string{
						"config-merger.k8s.io/watch":  "true",
						"config-merger.k8s.io/target": "merged-config",
					},
				},
				Data: map[string]string{
					"key2": "value2",
				},
			},
			expectedMerged: map[string]string{
				"config1.key1": "value1",
				"config2.key2": "value2",
			},
			expectCreation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			clientset := fake.NewSimpleClientset()

			// Create existing configmaps
			for _, cm := range tt.existingCMs {
				_, err := clientset.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Error creating test configmap: %v", err)
				}
			}

			// Create handler
			handler := NewConfigMapHandler(clientset)

			// Test the handler
			handler.HandleConfigMapChange(tt.inputCM)

			// Verify the result
			targetName := tt.inputCM.Annotations["config-merger.k8s.io/target"]
			merged, err := clientset.CoreV1().ConfigMaps(tt.inputCM.Namespace).Get(context.TODO(), targetName, metav1.GetOptions{})

			if tt.expectCreation {
				if err != nil {
					t.Errorf("Expected merged configmap to be created, but got error: %v", err)
				}
				if !reflect.DeepEqual(merged.Data, tt.expectedMerged) {
					t.Errorf("Merged data = %v, want %v", merged.Data, tt.expectedMerged)
				}
			}
		})
	}
}
