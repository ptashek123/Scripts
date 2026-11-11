package main

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Clean(u *unstructured.Unstructured) {
	paths := [][]string{
		{"metadata", "resourceVersion"},
		{"metadata", "uid"},
		{"metadata", "selfLink"},
		{"metadata", "creationTimestamp"},
		{"metadata", "deletionTimestamp"},
		{"metadata", "deletionGracePeriodSeconds"},
		{"metadata", "finalizers"},
		{"metadata", "ownerReferences"},
		{"metadata", "managedFields"},
		{"metadata", "annotations", "kubectl.kubernetes.io/last-applied-configuration"},
		{"status"},
	}

	for _, p := range paths {
		unstructured.RemoveNestedField(u.Object, p...)
	}
}