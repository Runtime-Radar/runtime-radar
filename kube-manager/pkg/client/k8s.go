package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kubernetes struct {
	client *kubernetes.Clientset
}

func NewKubernetes(cfg *rest.Config) (*Kubernetes, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can't create k8s client: %w", err)
	}

	return &Kubernetes{client}, nil
}

func (k *Kubernetes) DeletePod(ctx context.Context, namespace, name string, force bool) error {
	var gracePeriod *int64
	if force {
		gp := int64(0)
		gracePeriod = &gp
	}

	opts := metav1.DeleteOptions{
		GracePeriodSeconds: gracePeriod,
	}

	if err := k.client.CoreV1().Pods(namespace).Delete(ctx, name, opts); err != nil {
		return fmt.Errorf("can't delete pod: %w", err)
	}

	return nil
}
