package state

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/metrics"
)

type State int

const (
	Central State = iota
	ChildUnregistered
	ChildRegistered
	// If you change the State, don't forget to change metrics too
)

var instance struct {
	sync.RWMutex
	state State
}

func Get() State {
	instance.RLock()
	defer instance.RUnlock()
	return instance.state
}

func Set(v State) {
	instance.Lock()
	defer instance.Unlock()

	instance.state = v

	metrics.ClusterStateGauge.Set(float64(v))
	log.Debug().Stringer("state", v).Msg("State changed")
}

func (s State) String() string {
	switch s {
	case Central:
		return "Central"
	case ChildUnregistered:
		return "ChildUnregistered"
	case ChildRegistered:
		return "ChildRegistered"
	default:
		return "Unknown"
	}
}
