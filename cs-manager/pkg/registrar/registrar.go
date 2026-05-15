package registrar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	cluster_api "github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/model"
	"github.com/runtime-radar/runtime-radar/cs-manager/pkg/state"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var (
	errAlreadyRegistered = errors.New("child cluster already registered")
)

type Registrar struct {
	Interval               time.Duration
	Token                  uuid.UUID
	TokenHash              string
	ClusterController      cluster_api.ClusterControllerClient
	RegistrationRepository database.RegistrationRepository
}

func (r *Registrar) Run(stop <-chan struct{}) {
	log.Debug().Msgf("Registrar started")
	defer log.Debug().Msgf("Registrar stopped")

	t := time.NewTicker(r.Interval)
	defer t.Stop()

	for {
		// try to register instantly and then every tick
		reg, err := r.tryRegister(context.Background())
		switch {
		case errors.Is(err, errAlreadyRegistered):
			state.Set(state.ChildRegistered)
			log.Warn().Msgf("Child cluster already registered")
			return
		case err != nil:
			log.Error().Err(err).Interface("registration", reg).Msgf("Can't register child cluster")
		default:
			state.Set(state.ChildRegistered)
			log.Info().Interface("registration", reg).Msgf("Child cluster registered")
			return
		}

		select {
		case <-t.C:
			continue
		case <-stop:
			return
		}
	}
}

func (r *Registrar) tryRegister(ctx context.Context) (reg *model.Registration, err error) {
	// We should prevent this function from stopping whole program because of some unrecovered panic.
	defer func() {
		if p := recover(); p != nil {
			if errPanic, ok := p.(error); ok {
				err = errPanic
			} else if str, ok := p.(string); ok {
				err = errors.New(str)
			} else {
				panic(p) // something very special happened
			}
		}
	}()

	// We should check that there is an existing registration in case if there are multiple replicas of cs-manager
	reg, err = r.RegistrationRepository.GetLastSuccessful(ctx, r.TokenHash, false)
	if err == nil {
		return nil, fmt.Errorf("can't get registration: %w", err)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errAlreadyRegistered
	}

	reg = &model.Registration{TokenHash: r.TokenHash}

	registerReq := &cluster_api.RegisterClusterReq{
		Token: r.Token.String(),
	}
	_, err = r.ClusterController.Register(ctx, registerReq)
	if st, ok := status.FromError(err); ok && st.Code() == codes.Canceled {
		reg.Status = model.RegistrationStatusOK
	} else if err != nil {
		reg.Status = model.RegistrationStatusError
		reg.Error = err.Error()
		err = fmt.Errorf("can't register cluster: %w", err)
	} else {
		reg.Status = model.RegistrationStatusOK
	}

	// If synchronization with central CS was successful, save registration model to db
	if err := r.RegistrationRepository.Add(ctx, reg); err != nil {
		return nil, fmt.Errorf("can't save registration: %w", err)
	}

	return
}
