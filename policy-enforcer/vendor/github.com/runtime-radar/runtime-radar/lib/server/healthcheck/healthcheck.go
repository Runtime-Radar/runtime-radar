package healthcheck

import (
	"net/http"
	"sync"
)

var ready struct {
	sync.RWMutex
	ok bool
}

// IsReady checks app ready status.
func IsReady() bool {
	ready.RLock()
	defer ready.RUnlock()

	return ready.ok
}

// SetReady sets app ready status.
func SetReady() {
	ready.Lock()
	defer ready.Unlock()

	ready.ok = true
}

// ReadyHandler handles ready probes.
func ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	if IsReady() {
		// Make it simple: do nothing and have 200 OK
		return
	}

	http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
}

// LiveHandler handles live probes.
func LiveHandler(http.ResponseWriter, *http.Request) {
	// Make it simple: do nothing and have 200 OK
}
