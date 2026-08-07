package convert

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodToAPI(t *testing.T) {
	nilPod, err := PodToAPI(nil)
	require.NoError(t, err)
	require.Nil(t, nilPod)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "ns-1"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		Spec:       corev1.PodSpec{NodeName: "node-1", Containers: []corev1.Container{{Name: "c1"}}},
	}
	out, err := PodToAPI(pod)
	require.NoError(t, err)
	require.Equal(t, "pod-1", out.GetMetadata().GetName())
	require.Equal(t, "ns-1", out.GetMetadata().GetNamespace())
	require.Equal(t, "node-1", out.GetSpec().GetNodeName())
	require.Equal(t, string(corev1.PodRunning), out.GetStatus().GetPhase())
	require.Len(t, out.GetSpec().GetContainers(), 1)
	require.Equal(t, "c1", out.GetSpec().GetContainers()[0].GetName())
}

func TestNodeToAPI(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
			{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
		}},
	}
	out, err := NodeToAPI(node)
	require.NoError(t, err)
	require.Equal(t, "node-1", out.GetMetadata().GetName())
	require.Len(t, out.GetStatus().GetAddresses(), 1)
	require.Equal(t, "10.0.0.1", out.GetStatus().GetAddresses()[0].GetAddress())

	nilNode, err := NodeToAPI(nil)
	require.NoError(t, err)
	require.Nil(t, nilNode)
}

func TestPodsToAPI(t *testing.T) {
	pods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "a"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "b"}},
	}
	out, err := PodsToAPI(pods)
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "a", out[0].GetMetadata().GetName())
	require.Equal(t, "b", out[1].GetMetadata().GetName())
}
