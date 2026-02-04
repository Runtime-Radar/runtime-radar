//go:build !tinygo.wasm

package processor

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	detector_api "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/lib/security"
	notifier_api "github.com/runtime-radar/runtime-radar/notifier/api"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
)

const (
	minPoolSize   = 2
	reportTimeout = 10 * time.Millisecond
)

type Processor interface {
	Jobs() chan<- *tetragon.GetEventsResponse
	UpdateDetectors(bins [][]byte)
	Reports() <-chan *Report // for testing
	Config() *model.Config
	SetConfig(cfg *model.Config)
	Bins() (bins [][]byte, rootHash string)
	SetBins(bins [][]byte)
	Close()
}

type WorkersPool struct {
	history rabbit.PublishConsumer
	plugin  *detector_api.DetectorPlugin

	enforcer enforcer_api.EnforcerClient
	notifier notifier_api.NotifierClient

	jobs chan *tetragon.GetEventsResponse
	fire chan struct{}
	wg   *sync.WaitGroup

	updates []chan bool

	bins         [][]byte
	binsRootHash string
	binsMu       sync.RWMutex

	config   *model.Config
	configMu sync.RWMutex

	withReports bool         // for testing
	reports     chan *Report // for testing
}

type Report struct {
	ID     int
	Result any
	Err    error
	Delay  time.Duration
}

type WorkersPoolOption func(*WorkersPool)

// WithReports option enables reporting job results through a channel, available over Reports method. It's used in tests only.
func WithReports() WorkersPoolOption {
	return func(wp *WorkersPool) {
		wp.withReports = true
		wp.reports = make(chan *Report)
	}
}

func NewWorkersPool(
	poolSize, bufferSize int,
	history rabbit.PublishConsumer,
	plugin *detector_api.DetectorPlugin,
	enforcer enforcer_api.EnforcerClient,
	notifier notifier_api.NotifierClient,
	bins [][]byte,
	cfg *model.Config,
	opts ...WorkersPoolOption,
) (*WorkersPool, error) {
	wp := &WorkersPool{
		history: history,
		plugin:  plugin,

		enforcer: enforcer,
		notifier: notifier,

		bins:         bins,
		binsRootHash: BinsRootHashAsHex(bins),
		config:       cfg,

		jobs: make(chan *tetragon.GetEventsResponse, bufferSize),
		fire: make(chan struct{}),

		wg: &sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(wp)
	}

	if poolSize == 0 {
		// This should be set to correct value by automaxprocs lib
		poolSize = runtime.GOMAXPROCS(-1)
	}

	if poolSize < minPoolSize {
		poolSize = minPoolSize
	}

	wp.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		go wp.worker(i)
	}

	return wp, nil
}

func (wp *WorkersPool) Jobs() chan<- *tetragon.GetEventsResponse {
	return wp.jobs
}

func (wp *WorkersPool) Reports() <-chan *Report {
	if !wp.withReports {
		panic(fmt.Errorf("workers pool was not initialized with reports"))
	}

	return wp.reports
}

func (wp *WorkersPool) Close() {
	close(wp.fire)

	wp.wg.Wait()
}

// Config returns current config. It's safe for concurrent use.
func (wp *WorkersPool) Config() *model.Config {
	wp.configMu.RLock()
	defer wp.configMu.RUnlock()

	return wp.config
}

// SetConfig sets new config (but does not apply it). It's safe for concurrent use.
func (wp *WorkersPool) SetConfig(cfg *model.Config) {
	wp.configMu.Lock()
	defer wp.configMu.Unlock()

	wp.config = cfg
}

func (wp *WorkersPool) UpdateDetectors(bins [][]byte) {
	wp.binsMu.Lock()
	defer wp.binsMu.Unlock()

	wp.binsRootHash = BinsRootHashAsHex(bins)
	wp.bins = bins

	for _, upd := range wp.updates {
		select {
		case upd <- true:
		default:
			// Do nothing. If sending blocked, there is already update pending.
		}
	}
}

func (wp *WorkersPool) Bins() ([][]byte, string) {
	wp.binsMu.RLock()
	defer wp.binsMu.RUnlock()

	return wp.bins, wp.binsRootHash
}

func (wp *WorkersPool) SetBins(bins [][]byte) {
	wp.binsMu.Lock()
	defer wp.binsMu.Unlock()

	wp.binsRootHash = BinsRootHashAsHex(bins)
	wp.bins = bins
}

func (wp *WorkersPool) registerUpdate(upd chan bool) {
	wp.binsMu.Lock()
	defer wp.binsMu.Unlock()

	wp.updates = append(wp.updates, upd)
}

func BinsRootHashAsHex(bins [][]byte) string {
	hashes := make([][]byte, 0, len(bins))
	for _, b := range bins {
		hashes = append(hashes, security.HashSHA512(b))
	}

	return HashesHashAsHex(hashes)
}

// HashesHashAsHex returns hex-encoded SHA-512 hash of concatenated SHA-512 hashes.
// It takes non-encoded hashes as an argument.
func HashesHashAsHex(hashes [][]byte) string {
	bs := []byte{}
	for _, h := range hashes {
		bs = append(bs, h...)
	}

	return security.HashSHA512AsHex(bs)
}
