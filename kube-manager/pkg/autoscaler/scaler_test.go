package autoscaler

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
)

const testNamespace = "runtime-radar"

func testDeployment(replicas int32) *appsv1.Deployment {
	return namedDeployment("event-processor", replicas)
}

func namedDeployment(name string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  testNamespace,
			Generation: 1,
		},
		Spec: appsv1.DeploymentSpec{Replicas: ptr.To(replicas)},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas:  replicas,
			ObservedGeneration: 1,
		},
	}
}

func countUpdates(t *testing.T, client *fake.Clientset) int {
	t.Helper()

	var n int

	for _, action := range client.Actions() {
		if action.Matches("update", "deployments") {
			n++
		}
	}

	return n
}

func TestScalerGet(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	scaler := &Scaler{Client: client, Namespace: testNamespace}

	dep, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)
	require.Equal(t, int32(2), ptr.Deref(dep.Spec.Replicas, 0))

	_, err = scaler.Get(t.Context(), "missing")
	require.True(t, apierrors.IsNotFound(err), "Get(missing) error = %v, want NotFound", err)
}

func TestScalerApply(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	recorder := record.NewFakeRecorder(1)
	scaler := &Scaler{Client: client, Recorder: recorder, Namespace: testNamespace}

	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	dep, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	require.NoError(t, scaler.Apply(t.Context(), dep, 3, now, "backlog 1432 above 1000 for 4 samples"))

	stored, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	require.Equal(t, int32(3), ptr.Deref(stored.Spec.Replicas, 0))
	require.Equal(t, "2026-09-09T12:00:00Z", stored.Annotations[LastScaledAnnotation])
	require.True(t, LastScaled(stored).Equal(now), "LastScaled() = %v, want %v", LastScaled(stored), now)

	select {
	case got := <-recorder.Events:
		require.Equal(t, "Normal Scaled autoscaler: 2 -> 3 replicas, backlog 1432 above 1000 for 4 samples", got)
	default:
		t.Error("no event recorded")
	}
}

func TestScalerApplyAnnotatesWithoutAnnotations(t *testing.T) {
	dep := testDeployment(2)
	dep.Annotations = nil

	client := fake.NewClientset(dep)
	scaler := &Scaler{Client: client, Namespace: testNamespace}

	current, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	require.NoError(t, scaler.Apply(t.Context(), current, 1, time.Now(), "replicas 2 above max 1"))

	stored, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)
	require.Contains(t, stored.Annotations, LastScaledAnnotation)
}

func TestScalerApplyConflict(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	client.PrependReactor("update", "deployments", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewConflict(
			schema.GroupResource{Group: "apps", Resource: "deployments"},
			"event-processor",
			errors.New("the object has been modified"),
		)
	})

	recorder := record.NewFakeRecorder(1)
	scaler := &Scaler{Client: client, Recorder: recorder, Namespace: testNamespace}

	dep, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	err = scaler.Apply(t.Context(), dep, 3, time.Now(), "backlog high")
	require.ErrorIs(t, err, ErrConflict)
	require.Equal(t, 1, countUpdates(t, client), "the update was retried")

	select {
	case got := <-recorder.Events:
		t.Errorf("event recorded after a failed update: %q", got)
	default:
	}
}

func TestScalerApplyError(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	client.PrependReactor("update", "deployments", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewInternalError(errors.New("boom"))
	})

	scaler := &Scaler{Client: client, Namespace: testNamespace}

	dep, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	err = scaler.Apply(t.Context(), dep, 3, time.Now(), "backlog high")
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrConflict)
}

func TestScalerApplyNilRecorder(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	scaler := &Scaler{Client: client, Namespace: testNamespace}

	dep, err := scaler.Get(t.Context(), "event-processor")
	require.NoError(t, err)

	require.NoError(t, scaler.Apply(t.Context(), dep, 3, time.Now(), "backlog high"))
}

func TestLastScaled(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		dep  *appsv1.Deployment
		want time.Time
	}{
		{
			name: "no annotations",
			dep:  &appsv1.Deployment{},
			want: time.Time{},
		},
		{
			name: "unrelated annotations",
			dep: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"other": "value"}},
			},
			want: time.Time{},
		},
		{
			name: "unparsable value",
			dep: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{LastScaledAnnotation: "yesterday"}},
			},
			want: time.Time{},
		},
		{
			name: "valid value",
			dep: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{LastScaledAnnotation: "2026-09-09T12:00:00Z"},
				},
			},
			want: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastScaled(tt.dep)
			require.True(t, got.Equal(tt.want), "LastScaled() = %v, want %v", got, tt.want)
		})
	}
}

func TestObservationFrom(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Generation:  7,
			Annotations: map[string]string{LastScaledAnnotation: "2026-09-09T12:00:00Z"},
		},
		Spec: appsv1.DeploymentSpec{Replicas: ptr.To[int32](3)},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas:  2,
			ObservedGeneration: 6,
		},
	}

	require.Equal(t, Observation{
		Backlog:            42,
		Replicas:           3,
		Available:          2,
		Generation:         7,
		ObservedGeneration: 6,
		LastScaled:         now,
	}, ObservationFrom(dep, 42))
}

func TestObservationFromNilReplicas(t *testing.T) {
	dep := &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: nil}}

	got := ObservationFrom(dep, 0)
	require.Equal(t, int32(1), got.Replicas)
	require.True(t, got.LastScaled.IsZero(), "LastScaled = %v, want zero", got.LastScaled)
}
