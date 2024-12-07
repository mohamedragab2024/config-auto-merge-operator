package controller

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestConfigMapController(t *testing.T) {
	// Create fake clientset
	clientset := fake.NewSimpleClientset()

	// Create controller
	controller := NewConfigMapController(clientset)

	// Create test configmap
	testCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
			Annotations: map[string]string{
				"config-merger.k8s.io/watch":  "true",
				"config-merger.k8s.io/target": "merged-config",
			},
		},
		Data: map[string]string{
			"key1": "value1",
		},
	}

	// Start the controller
	stopCh := make(chan struct{})
	defer close(stopCh)
	go controller.Run(stopCh)

	// Wait for cache sync
	if !cache.WaitForCacheSync(stopCh, controller.informer.HasSynced) {
		t.Fatal("Timed out waiting for caches to sync")
	}

	// Test Add
	t.Run("Test Add ConfigMap", func(t *testing.T) {
		controller.handleAdd(testCM)
		// Verify the merged configmap was created
		merged, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "merged-config", metav1.GetOptions{})
		if err != nil {
			t.Errorf("Expected merged configmap to be created, but got error: %v", err)
		}
		expectedData := map[string]string{
			"test-config.key1": "value1",
		}
		if !reflect.DeepEqual(merged.Data, expectedData) {
			t.Errorf("Merged data = %v, want %v", merged.Data, expectedData)
		}
	})

	// Test Update
	t.Run("Test Update ConfigMap", func(t *testing.T) {
		updatedCM := testCM.DeepCopy()
		updatedCM.Data["key1"] = "updated-value"
		controller.handleUpdate(testCM, updatedCM)

		merged, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "merged-config", metav1.GetOptions{})
		if err != nil {
			t.Errorf("Error getting merged configmap: %v", err)
		}
		expectedData := map[string]string{
			"test-config.key1": "updated-value",
		}
		if !reflect.DeepEqual(merged.Data, expectedData) {
			t.Errorf("Merged data = %v, want %v", merged.Data, expectedData)
		}
	})

	// Test Delete
	t.Run("Test Delete ConfigMap", func(t *testing.T) {
		controller.handleDelete(testCM)
		merged, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "merged-config", metav1.GetOptions{})
		if err != nil {
			t.Errorf("Error getting merged configmap: %v", err)
		}
		if len(merged.Data) != 0 {
			t.Errorf("Expected empty merged data, got %v", merged.Data)
		}
	})
}
