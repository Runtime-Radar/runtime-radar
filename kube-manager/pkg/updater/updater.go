package updater

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/wait"
)

//go:generate minimock -i ConfigUpdater -o . -s _mock.go -g
type ConfigUpdater interface {
	Config() *model.Config
	SetConfig(cfg *model.Config)
	Maps() (map[string]struct{}, map[string]struct{})
	Equal(oldCfg, newCfg *model.Config) bool
	SetOnUpdateFunc(f func() error)
}

type Service struct {
	config           *model.Config
	updateInterval   time.Duration
	configRepository database.ConfigRepository

	onUpdate func() error

	m sync.RWMutex
}

func New(ctx context.Context, db *gorm.DB, updateInterval time.Duration) (*Service, error) {
	s := &Service{
		updateInterval:   updateInterval,
		configRepository: &database.ConfigDatabase{DB: db},
		config:           model.DefaultConfig,
	}

	err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) SetOnUpdateFunc(f func() error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.onUpdate = f
}

// Config returns current Tetra config. It's safe for concurrent use.
func (s *Service) Config() *model.Config {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.config
}

func (s *Service) SetConfig(cfg *model.Config) {
	if s.onUpdate != nil {
		err := s.onUpdate()
		if err != nil {
			log.Error().Err(err).Msg("error executing onUpdate function")
		}
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.config = cfg
}

func (s *Service) Maps() (map[string]struct{}, map[string]struct{}) {
	cfg := s.Config()

	allow := make(map[string]struct{})
	for _, c := range cfg.AllowNamespaces() {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		allow[c] = struct{}{}
	}

	deny := make(map[string]struct{})
	for _, c := range cfg.DenyNamespaces() {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		deny[c] = struct{}{}
	}

	return allow, deny
}

func (s *Service) Run(stop chan struct{}) {
	ctx := wait.ContextForChannel(stop)

	log.Info().Msgf("Filter config updater started")
	defer log.Info().Msgf("Filter config updater stopped")

	t := time.NewTicker(s.updateInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := s.loadConfig(ctx); err != nil {
				log.Error().Err(err).Msg("can't load config from DB")
			}

		case <-stop:
			return
		}
	}
}

func (s *Service) Equal(oldCfg, newCfg *model.Config) bool {
	return cmp.Equal(oldCfg.AllowNamespaces(), newCfg.AllowNamespaces(), protocmp.Transform()) &&
		cmp.Equal(oldCfg.DenyNamespaces(), newCfg.DenyNamespaces(), protocmp.Transform())
}

func (s *Service) loadConfig(ctx context.Context) error {
	log.Debug().Msg("configuration update via DB has started")
	newCfg, err := s.configRepository.GetLast(ctx, true)
	if err != nil {
		return err
	}
	oldCfg := s.Config()

	if s.Equal(oldCfg, newCfg) {
		log.Debug().Msg("configuration has no changes")
		return nil
	}
	s.SetConfig(newCfg)
	log.Debug().Msg("configuration has been updated")

	return nil
}
