package utils

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func AddNamespaces(ctx context.Context, fakeClient *fake.Clientset, nss ...string) error {
	for _, nsName := range nss {
		ns := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				UID:  types.UID(nsName + "-namespace-uid"),
				Name: nsName,
			},
			Spec:   corev1.NamespaceSpec{},
			Status: corev1.NamespaceStatus{},
		}

		_, err := fakeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteNamespaces(ctx context.Context, fakeClient *fake.Clientset, nss ...string) error {
	for _, nsName := range nss {
		err := fakeClient.CoreV1().Namespaces().Delete(ctx, nsName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func AddPods(ctx context.Context, fakeClient *dynamicfake.FakeDynamicClient, pods ...*corev1.Pod) error {
	for _, pod := range pods {
		unstructuredPod, err := convertToUnstructured(pod)
		if err != nil {
			return err
		}

		_, err = fakeClient.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		}).Namespace(pod.Namespace).Create(ctx, unstructuredPod, metav1.CreateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func AddNodes(ctx context.Context, fakeClient *dynamicfake.FakeDynamicClient, nodes ...*corev1.Node) error {
	for _, n := range nodes {
		u, err := convertToUnstructured(n)
		if err != nil {
			return err
		}

		_, err = fakeClient.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "nodes",
		}).Namespace("").Create(ctx, u, metav1.CreateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func UpdatePods(ctx context.Context, fakeClient *dynamicfake.FakeDynamicClient, pods ...*corev1.Pod) error {
	for _, pod := range pods {
		unstructuredPod, err := convertToUnstructured(pod)
		if err != nil {
			return err
		}

		_, err = fakeClient.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		}).Namespace(pod.Namespace).Update(ctx, unstructuredPod, metav1.UpdateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func DeletePods(ctx context.Context, fakeClient *dynamicfake.FakeDynamicClient, pods ...*corev1.Pod) error {
	for _, pod := range pods {
		err := fakeClient.Resource(schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		}).Namespace(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func convertToUnstructured(object interface{}) (*unstructured.Unstructured, error) {
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: o}, nil
}
