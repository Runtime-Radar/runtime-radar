package autoscaler

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// LeaseName is constant: two autoscalers in one namespace would fight over the same Deployments.
const LeaseName = "kube-manager-autoscaler"

// leaderTimings are the conventional Kubernetes controller values; a var so tests can shorten
// the election.
var leaderTimings = struct {
	Lease, Renew, Retry time.Duration
}{
	Lease: 15 * time.Second,
	Renew: 10 * time.Second,
	Retry: 2 * time.Second,
}

// RunLeader calls run whenever this replica holds the Lease, with a context cancelled on loss,
// and re-enters the election after every lost term until ctx is cancelled. A term is joined
// before the next one begins, so run never overlaps with itself. identity should be the pod
// hostname: it is stable across restarts, so a restarted pod reclaims its own Lease.
func RunLeader(ctx context.Context, client kubernetes.Interface, namespace, identity string, run func(ctx context.Context)) error {
	if identity == "" {
		return errors.New("leader election identity is empty")
	}

	config := leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta:  metav1.ObjectMeta{Name: LeaseName, Namespace: namespace},
			Client:     client.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{Identity: identity},
		},
		LeaseDuration:   leaderTimings.Lease,
		RenewDeadline:   leaderTimings.Renew,
		RetryPeriod:     leaderTimings.Retry,
		ReleaseOnCancel: true, // so the surviving replica takes over at once
		Name:            LeaseName,
		Callbacks: leaderelection.LeaderCallbacks{
			// Fires on every elector return, even for a candidate that never led; the term
			// callback below reports the release instead.
			OnStoppedLeading: func() {},
			OnNewLeader: func(leader string) {
				if leader != identity {
					log.Debug().Str("leader", leader).Msg("Another replica holds the autoscaler lease")
				}
			},
		},
	}

	for {
		if ctx.Err() != nil {
			return nil
		}

		current := newTerm()
		config.Callbacks.OnStartedLeading = func(ctx context.Context) {
			if !current.start() {
				return
			}
			defer current.finish()

			defer func() {
				metrics.AutoscalerIsLeader.Set(0)
				log.Info().Str("identity", identity).Msg("Released the autoscaler lease")
			}()

			metrics.AutoscalerIsLeader.Set(1)
			log.Info().Str("identity", identity).Msg("Acquired the autoscaler lease")
			run(ctx)
		}

		elector, err := leaderelection.NewLeaderElector(config)
		if err != nil {
			return fmt.Errorf("build leader elector for lease %s/%s: %w", namespace, LeaseName, err)
		}

		elector.Run(ctx)
		current.wait()

		if ctx.Err() != nil {
			return nil
		}

		log.Info().Str("identity", identity).Msg("Lost the autoscaler lease, re-entering the election")

		select {
		case <-ctx.Done():
			return nil

		case <-time.After(leaderTimings.Retry):
		}
	}
}

const (
	termPending int32 = iota
	termRunning
	termAbandoned
)

// term joins the run callback of one leadership term: client-go starts it in its own goroutine
// and never waits for it, so a regained lease could otherwise start a second run mid-tick.
type term struct {
	state atomic.Int32
	done  chan struct{}
}

func newTerm() *term {
	return &term{done: make(chan struct{})}
}

// start returns false once wait has abandoned the term, i.e. the elector returned before
// client-go scheduled the callback at all.
func (t *term) start() bool {
	return t.state.CompareAndSwap(termPending, termRunning)
}

func (t *term) finish() {
	close(t.done)
}

func (t *term) wait() {
	if t.state.CompareAndSwap(termPending, termAbandoned) {
		return
	}

	<-t.done
}
