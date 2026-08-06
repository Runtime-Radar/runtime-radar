package inventory

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/config"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/updater"
	testutils "github.com/runtime-radar/runtime-radar/kube-manager/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

func TestService_WithConnect(t *testing.T) {
	cfg := config.New()

	if cfg.KubeConfig == "" {
		t.Skip("KUBECONFIG env var not set")
	}

	mc := minimock.NewController(t)
	updaterMock := updater.NewConfigUpdaterMock(mc)
	updaterMock.MapsMock.Return(map[string]struct{}{}, map[string]struct{}{})
	updaterMock.SetOnUpdateFuncMock.Return()

	stopCh := make(chan struct{})

	// use the current context in kube config
	k8sCfg, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	require.NoError(t, err)

	podInf := informers.New[*corev1.Pod]("pods")
	nodeInf := informers.New[*corev1.Node]("nodes")

	clientset, err := kubernetes.NewForConfig(k8sCfg)
	require.NoError(t, err)

	dynamicClient, err := dynamic.NewForConfig(k8sCfg)
	require.NoError(t, err)

	resource := &resourceClient{}
	inv, err := NewWithClients(updaterMock, clientset, dynamicClient, resource, 0, podInf, nodeInf)
	require.NoError(t, err)

	err = inv.Run(stopCh)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// resources
	list, err := resource.Get(clientset.Discovery())
	require.NoError(t, err)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Resource < list[j].Resource
	})

	for idx, r := range list {
		t.Logf("%d resource: {'%s', '%s', '%s'}", idx+1, r.Resource, r.Group, r.Version)
	}

	// pods
	pods, total := podInf.List()
	require.Len(t, pods, total)

	for idx, pod := range pods {
		get, err := podInf.Get(pod.Namespace, pod.Name)
		require.NoError(t, err)
		assert.Equal(t, pod, get)

		t.Logf("%d pod: {%s, %s, %s}", idx+1, get.GetUID(), get.GetNamespace(), get.GetName())
	}

	// nodes
	nodes, total := nodeInf.List()
	require.Len(t, nodes, total)

	for idx, node := range nodes {
		get, err := nodeInf.Get(defaultNamespace, node.Name)
		require.NoError(t, err)
		assert.Equal(t, node, get)

		t.Logf("%d node: {%s, %s, %s}", idx+1, get.GetUID(), get.GetNamespace(), get.GetName())
	}

	time.Sleep(100 * time.Millisecond)

	close(stopCh)
	inv.Shutdown()
}

func TestService_WithFake(t *testing.T) {
	ctx := context.Background()
	fakeClient := fake.NewClientset()

	tm1 := time.Now().Add(-time.Hour).Round(time.Second)
	tm2 := tm1.Add(time.Minute).Round(time.Second)
	tm3 := tm2.Add(time.Minute).Round(time.Second)
	pods := []*corev1.Pod{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("pod-1"),
				Name:              "pod-1",
				Namespace:         "ns-1",
				CreationTimestamp: metav1.Time{Time: tm1},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("pod-2"),
				Name:              "pod-2",
				Namespace:         "ns-2",
				CreationTimestamp: metav1.Time{Time: tm2},
			},
		}, {
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				UID:               types.UID("pod-3"),
				Name:              "pod-3",
				Namespace:         "ns-2",
				CreationTimestamp: metav1.Time{Time: tm3},
			},
		},
	}

	scheme := runtime.NewScheme()
	gv := schema.GroupVersion{Group: "", Version: "v1"}
	scheme.AddKnownTypes(gv, &corev1.Pod{})
	scheme.AddKnownTypes(gv, &corev1.PodList{})
	resources := []schema.GroupVersionResource{
		gv.WithResource("pods"),
	}
	podInformer := informers.New[*corev1.Pod]("pods")

	fakeDynamic := dynamicfake.NewSimpleDynamicClient(scheme)

	mc := minimock.NewController(t)
	updaterMock := updater.NewConfigUpdaterMock(mc)
	updaterMock.MapsMock.Return(map[string]struct{}{}, map[string]struct{}{})
	updaterMock.SetOnUpdateFuncMock.Return()

	stopCh := make(chan struct{})

	inv, err := NewWithClients(updaterMock, fakeClient, fakeDynamic, &testutils.FakeResourcer{Resources: resources}, 0, podInformer)
	require.NoError(t, err)

	err = testutils.AddNamespaces(ctx, fakeClient, "ns-1", "ns-2")
	require.NoError(t, err)

	err = testutils.AddPods(ctx, fakeDynamic, pods...)
	require.NoError(t, err)

	err = inv.Run(stopCh)
	require.NoError(t, err)

	waitListSizeOrTimeout(t, podInformer, 3, time.Second)

	update1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID("pod-1"),
			Name:              "pod-1",
			Namespace:         "ns-1",
			CreationTimestamp: metav1.Time{Time: tm1},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	// verify initial list
	act, err := podInformer.Get("ns-1", "pod-1")
	require.NoError(t, err)
	assert.Equal(t, pods[0], act)

	act, err = podInformer.Get("ns-2", "pod-2")
	require.NoError(t, err)
	assert.Equal(t, pods[1], act)

	act, err = podInformer.Get("ns-2", "pod-3")
	require.NoError(t, err)
	assert.Equal(t, pods[2], act)

	// update and delete pods
	err = testutils.UpdatePods(ctx, fakeDynamic, update1)
	require.NoError(t, err)

	err = testutils.DeletePods(ctx, fakeDynamic, pods[1])
	require.NoError(t, err)

	err = inv.onUpdate()
	require.NoError(t, err)

	waitListSizeOrTimeout(t, podInformer, 2, time.Second)

	act, err = podInformer.Get("ns-1", "pod-1")
	require.NoError(t, err)
	assert.Equal(t, update1, act)

	_, err = podInformer.Get("ns-2", "pod-2")
	require.Error(t, err)
	assert.Equal(t, informers.ErrNotFound, err)

	list, total := podInformer.List()
	assert.Len(t, list, total)
	assert.Equal(t, total, 2)
	defaultSort(list)

	assert.Equal(t, update1, list[0])
	assert.Equal(t, pods[2], list[1])

	// Add namespace
	err = testutils.AddNamespaces(ctx, fakeClient, "ns-3")
	require.NoError(t, err)

	pns3 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:               "uid-pod-1",
			Name:              "pod-1",
			Namespace:         "ns-3",
			CreationTimestamp: metav1.Time{Time: tm1},
		},
	}
	err = testutils.AddPods(ctx, fakeDynamic, pns3)
	require.NoError(t, err)

	waitListSizeOrTimeout(t, podInformer, 3, time.Second)

	act, err = podInformer.Get("ns-3", "pod-1")
	require.NoError(t, err)
	assert.Equal(t, pns3, act)

	list, total = podInformer.List()
	assert.Len(t, list, total)
	assert.Equal(t, total, 3)
	defaultSort(list)

	assert.Equal(t, update1, list[0])
	assert.Equal(t, pods[2], list[1])
	assert.Equal(t, pns3, list[2])

	// Remove namespace
	err = testutils.DeleteNamespaces(ctx, fakeClient, "ns-2")
	require.NoError(t, err)

	waitListSizeOrTimeout(t, podInformer, 2, time.Second)

	_, err = podInformer.Get("ns-2", "pod-2")
	require.Error(t, err)
	assert.Equal(t, informers.ErrNotFound, err)

	list, total = podInformer.List()
	assert.Len(t, list, total)
	assert.Equal(t, total, 2)
	defaultSort(list)

	assert.Equal(t, update1, list[0])
	assert.Equal(t, pns3, list[1])

	close(stopCh)
	inv.Shutdown()
}

func waitListSizeOrTimeout(t *testing.T, informer *informers.Informer[*corev1.Pod], size int, timeout time.Duration) {
	dur := timeout / 10
	for i := 0; i < 10; i++ {
		_, s := informer.List()
		if s != size {
			time.Sleep(dur)
			continue
		}

		return
	}

	assert.Fail(t, "timeout of informer get list")
}

func defaultSort(l []*corev1.Pod) {
	sort.Slice(l, func(i, j int) bool {
		if l[i].Namespace == l[j].Namespace {
			return l[i].Name < l[j].Name
		}
		return l[i].Namespace < l[j].Namespace
	})
}
