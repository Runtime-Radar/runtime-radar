package convert

import (
	"fmt"

	apicorev1 "github.com/runtime-radar/runtime-radar/kube-manager/api/k8s.io/api/core/v1"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

func PodToAPI(p *corev1.Pod) (*apicorev1.Pod, error) {
	if p == nil {
		return nil, nil
	}
	b, err := p.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal pod: %w", err)
	}
	out := new(apicorev1.Pod)
	if err := proto.Unmarshal(b, out); err != nil {
		return nil, fmt.Errorf("unmarshal pod: %w", err)
	}
	return out, nil
}

func PodsToAPI(pods []*corev1.Pod) ([]*apicorev1.Pod, error) {
	out := make([]*apicorev1.Pod, 0, len(pods))
	for _, p := range pods {
		converted, err := PodToAPI(p)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

func NodeToAPI(n *corev1.Node) (*apicorev1.Node, error) {
	if n == nil {
		return nil, nil
	}
	b, err := n.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal node: %w", err)
	}
	out := new(apicorev1.Node)
	if err := proto.Unmarshal(b, out); err != nil {
		return nil, fmt.Errorf("unmarshal node: %w", err)
	}
	return out, nil
}

func NodesToAPI(nodes []*corev1.Node) ([]*apicorev1.Node, error) {
	out := make([]*apicorev1.Node, 0, len(nodes))
	for _, n := range nodes {
		converted, err := NodeToAPI(n)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}
