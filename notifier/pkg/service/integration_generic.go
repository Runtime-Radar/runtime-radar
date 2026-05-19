package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/cipher"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/database"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model/convert"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/notifier"
	enforcer_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type IntegrationGeneric struct {
	api.UnimplementedIntegrationControllerServer

	IntegrationRepository  database.IntegrationRepository
	NotificationRepository database.NotificationRepository
	RuleController         enforcer_api.RuleControllerClient
	Crypter                cipher.Crypter
}

func (ig *IntegrationGeneric) Create(ctx context.Context, req *api.Integration) (*api.CreateIntegrationResp, error) {
	if reason, ok := validateIntegration(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	var (
		i   model.Integration
		err error
	)
	i, err = convert.IntegrationFromPB(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse integration: %v", err)
	}

	if !req.GetSkipCheck() {
		n, err := notifier.FromIntegration(i)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if err := n.Test(ctx); err != nil {
			msg := fmt.Sprintf("integration is inaccessible: %v", err)
			return nil, errcommon.StatusWithReason(codes.InvalidArgument, IntegrationInaccessible, msg).Err()
		}
	}

	// encrypt sensitive data before persisting to storage
	i.EncryptSensitive(ig.Crypter)

	if err := ig.IntegrationRepository.Add(ctx, i); err != nil {
		if errors.Is(err, model.ErrIntegrationNameInUse) {
			return nil, errcommon.StatusWithReason(codes.AlreadyExists, NameMustBeUnique, "name must be unique").Err()
		}

		return nil, status.Errorf(codes.Internal, "can't save integration: %v", err)
	}

	return &api.CreateIntegrationResp{Id: i.GetID().String()}, nil
}

func (ig *IntegrationGeneric) Read(ctx context.Context, req *api.ReadIntegrationReq) (*api.Integration, error) {
	if reason, ok := validateIntegrationType(req.GetType()); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is not set")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	i, err := ig.IntegrationRepository.GetByTypeAndID(ctx, req.GetType(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "integration not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get integration: %v", err)
	}

	return convert.IntegrationToPB(i), nil
}

func (ig *IntegrationGeneric) Update(ctx context.Context, req *api.Integration) (*emptypb.Empty, error) {
	if reason, ok := validateIntegration(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is not set")
	}

	i, err := convert.IntegrationFromPB(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse integration: %v", err)
	}

	if !req.GetSkipCheck() {
		n, err := notifier.FromIntegration(i)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if err := n.Test(ctx); err != nil {
			msg := fmt.Sprintf("integration is inaccessible: %v", err)
			return nil, errcommon.StatusWithReason(codes.InvalidArgument, IntegrationInaccessible, msg).Err()
		}
	}

	var updateMap map[string]any

	switch conf := req.GetConfig().(type) {
	case *api.Integration_Email:
		updateMap = ig.emailUpdateMap(req.GetName(), conf.Email)
	case *api.Integration_Webhook:
		updateMap = ig.webhookUpdateMap(req.GetName(), conf.Webhook)
	case *api.Integration_Syslog:
		updateMap = ig.syslogUpdateMap(req.GetName(), conf.Syslog)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid config type given: %T", conf)
	}

	if err := ig.IntegrationRepository.UpdateWithMap(ctx, req.GetType(), i.GetID(), updateMap); err != nil {
		if errors.Is(err, model.ErrIntegrationNameInUse) {
			return nil, errcommon.StatusWithReason(codes.AlreadyExists, NameMustBeUnique, "name must be unique").Err()
		}

		return nil, status.Errorf(codes.Internal, "can't update integration: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (ig *IntegrationGeneric) Delete(ctx context.Context, req *api.DeleteIntegrationReq) (*emptypb.Empty, error) {
	if reason, ok := validateIntegrationType(req.GetType()); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is not set")
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	notifications, err := ig.NotificationRepository.GetByIntegrationID(ctx, id, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get integration's notifications: %v", err)
	}

	if len(notifications) > 0 {
		rcResp, err := ig.RuleController.NotifyTargetsInUse(ctx, &enforcer_api.NotifyTargetsInUseReq{
			Targets: idsFromNotifications(notifications...),
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can't check if notifications are in use: %v", err)
		}

		if rcResp.InUse {
			return nil, errcommon.StatusWithReason(codes.FailedPrecondition, NotificationInUse, "integration's notifications are in use by rule").Err()
		}
	}

	if err := ig.IntegrationRepository.DeleteByTypeAndID(ctx, req.GetType(), id); err != nil {
		return nil, status.Errorf(codes.Internal, "can't delete integration: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (ig *IntegrationGeneric) List(ctx context.Context, req *api.ListIntegrationReq) (*api.ListIntegrationResp, error) {
	if reason, ok := validateIntegrationType(req.GetType()); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	order := req.GetOrder()
	if order == "" {
		order = defaultOrder
	}

	integrations, err := ig.IntegrationRepository.GetAllByType(ctx, req.GetType(), order)
	if err != nil {
		if errors.Is(err, database.ErrInvalidOrder) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "can't get integrations: %v", err)
	}

	return &api.ListIntegrationResp{
		Integrations: convert.IntegrationsToPB(integrations),
	}, nil
}

func validateIntegration(req *api.Integration) (reason string, valid bool) {
	if req.GetName() == "" {
		return "name is empty", false
	}
	if req.GetConfig() == nil {
		return "config is nil", false
	}

	switch req.GetType() {
	case model.IntegrationEmail:
		conf, ok := req.GetConfig().(*api.Integration_Email)
		if !ok {
			return fmt.Sprintf("integration type %s does not match config type %T", req.GetType(), req.GetConfig()), false
		}

		reason, valid = validateEmail(conf.Email)
		if !valid {
			return
		}

	case model.IntegrationWebhook:
		conf, ok := req.GetConfig().(*api.Integration_Webhook)
		if !ok {
			return fmt.Sprintf("integration type %s does not match config type %T", req.GetType(), req.GetConfig()), false
		}

		reason, valid = validateWebhook(conf.Webhook)
		if !valid {
			return
		}

	case model.IntegrationSyslog:
		conf, ok := req.GetConfig().(*api.Integration_Syslog)
		if !ok {
			return fmt.Sprintf("integration type %s does not match config type %T", req.GetType(), req.GetConfig()), false
		}

		reason, valid = validateSyslog(conf.Syslog)
		if !valid {
			return
		}

	default:
		return fmt.Sprintf("unsupported integration type given: %s", req.GetType()), false
	}

	return "", true
}

func validateIntegrationType(it string) (reason string, valid bool) {
	if it == "" {
		return "integration type is empty", false
	}
	if !model.IntegrationTypeSupported(it) {
		return fmt.Sprintf("unsupported integration type given: %s", it), false
	}
	return "", true
}

func validateEmail(e *api.Email) (reason string, valid bool) {
	if e.GetServer() == "" {
		return "server is empty", false
	}
	return "", true
}

func validateWebhook(w *api.Webhook) (reason string, valid bool) {
	if w.GetUrl() == "" {
		return "url is empty", false
	}
	return "", true
}

func validateSyslog(w *api.Syslog) (reason string, valid bool) {
	if w.GetAddress() == "" {
		return "address is empty", false
	}

	_, err := url.Parse(w.GetAddress())
	if err != nil {
		return fmt.Sprintf("can't parse address '%s': %+v", w.GetAddress(), err), false
	}

	return "", true
}

func idsFromNotifications(ns ...*model.Notification) []string {
	res := make([]string, 0, len(ns))
	for _, n := range ns {
		res = append(res, n.ID.String())
	}
	return res
}

func (ig *IntegrationGeneric) emailUpdateMap(name string, conf *api.Email) map[string]any {
	encryptedPassword := ""
	if conf.GetPassword() != "" {
		encryptedPassword = ig.Crypter.EncryptStringAsHex(conf.GetPassword())
	}

	return map[string]any{
		"Name":              name,
		"From":              conf.GetFrom(),
		"Server":            conf.GetServer(),
		"AuthType":          conf.GetAuthType(),
		"Username":          conf.GetUsername(),
		"EncryptedPassword": encryptedPassword,
		"UseTLS":            conf.GetUseTls(),
		"UseStartTLS":       conf.GetUseStartTls(),
		"Insecure":          conf.GetInsecure(),
		"CA":                conf.GetCa(),
	}
}

func (ig *IntegrationGeneric) webhookUpdateMap(name string, conf *api.Webhook) map[string]any {
	encryptedPassword := ""
	if conf.GetPassword() != "" {
		encryptedPassword = ig.Crypter.EncryptStringAsHex(conf.GetPassword())
	}

	return map[string]any{
		"Name":              name,
		"URL":               conf.GetUrl(),
		"Login":             conf.GetLogin(),
		"EncryptedPassword": encryptedPassword,
		"Insecure":          conf.GetInsecure(),
		"CA":                conf.GetCa(),
	}
}

func (ig *IntegrationGeneric) syslogUpdateMap(name string, conf *api.Syslog) map[string]any {
	return map[string]any{
		"Name":    name,
		"Address": conf.GetAddress(),
	}
}
