package autoscaler

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
)

// LastScaledAnnotation holds the RFC3339 time of the last action; it makes the cooldown survive a
// leadership change.
const LastScaledAnnotation = "runtime-radar.io/autoscaler-last-scaled-at"

var ErrConflict = errors.New("deployment update conflict")

// Scaler reads and updates the target Deployments.
type Scaler struct {
	Client    kubernetes.Interface
	Recorder  record.EventRecorder // optional
	Namespace string
}

// Get reads from the API server, not an informer: a cached resourceVersion turns updates into conflicts.
func (s *Scaler) Get(ctx context.Context, name string) (*appsv1.Deployment, error) {
	return s.Client.AppsV1().Deployments(s.Namespace).Get(ctx, name, metav1.GetOptions{})
}

// Apply sets spec.replicas and the last-scaled annotation in one update. It carries dep's
// resourceVersion, so a Deployment changed since Get fails with ErrConflict instead of being
// overwritten; the caller re-reads on the next tick rather than retrying.
func (s *Scaler) Apply(ctx context.Context, dep *appsv1.Deployment, desired int32, now time.Time, reason string) error {
	from := ptr.Deref(dep.Spec.Replicas, 1)

	dep.Spec.Replicas = ptr.To(desired)

	if dep.Annotations == nil {
		dep.Annotations = make(map[string]string, 1)
	}

	dep.Annotations[LastScaledAnnotation] = now.UTC().Format(time.RFC3339)

	if _, err := s.Client.AppsV1().Deployments(s.Namespace).Update(ctx, dep, metav1.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			return fmt.Errorf("%w: %s: %w", ErrConflict, dep.Name, err)
		}

		return fmt.Errorf("update deployment %s: %w", dep.Name, err)
	}

	if s.Recorder != nil {
		s.Recorder.Eventf(dep, corev1.EventTypeNormal, "Scaled", "autoscaler: %d -> %d replicas, %s", from, desired, reason)
	}

	return nil
}

func ObservationFrom(dep *appsv1.Deployment, backlog int) Observation {
	return Observation{
		Backlog:            backlog,
		Replicas:           ptr.Deref(dep.Spec.Replicas, 1),
		Available:          dep.Status.AvailableReplicas,
		Generation:         dep.Generation,
		ObservedGeneration: dep.Status.ObservedGeneration,
		LastScaled:         LastScaled(dep),
	}
}

// LastScaled returns the zero time for a missing or unparsable annotation, which disables the
// cooldown rather than blocking on it.
func LastScaled(dep *appsv1.Deployment) time.Time {
	ts, err := time.Parse(time.RFC3339, dep.Annotations[LastScaledAnnotation])
	if err != nil {
		return time.Time{}
	}

	return ts
}
