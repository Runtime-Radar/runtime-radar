package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/cipher"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/database"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/notifier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotifierGeneric struct {
	api.UnimplementedNotifierServer

	NotificationRepository database.NotificationRepository
	IntegrationRepository  database.IntegrationRepository
	Crypter                cipher.Crypter

	CSVersion string
}

// Notify tries to notify list of targets about events.
// It dispatches sender (Integration) and tries to send notification via it.
// In case when at least one Notification with given id does not exist, error is returned without notifying anyone.
// If at least one of notifications fail it doesn't block others, but error with reason NOTIFICATION_FAILED is returned.
func (ng *NotifierGeneric) Notify(ctx context.Context, req *api.NotifyReq) (*emptypb.Empty, error) {
	if reason, ok := ng.validateNotifyReq(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	ns := req.GetNotifications()
	ids := make([]uuid.UUID, 0, len(ns))

	// Collect all notification IDs to fetch them from DB with one query
	for _, msg := range ns {
		id, err := uuid.Parse(msg.GetNotificationId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
		}
		ids = append(ids, id)
	}

	notifications, err := ng.NotificationRepository.GetByIDs(ctx, ids, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get notifications: %v", err)
	}

	integrations, err := ng.IntegrationRepository.GetByNotifications(ctx, notifications...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get integrations: %v", err)
	}

	integrationsMap := make(map[uuid.UUID]model.Integration, len(integrations))
	for _, ig := range integrations {
		ig.SetMeta(model.IntegrationMeta{CSVersion: ng.CSVersion})
		integrationsMap[ig.GetID()] = ig
	}

	notificationsMap := make(map[uuid.UUID]*model.Notification, len(notifications))
	for _, n := range notifications {
		notificationsMap[n.ID] = n
	}

	for _, id := range ids {
		if notificationsMap[id] == nil {
			return nil, status.Errorf(codes.InvalidArgument, "notification %s not found", id)
		}
	}

	errs := []error{}

	for i, msg := range ns {
		// no need to check existence because we did it above
		notification := notificationsMap[ids[i]]

		ig := integrationsMap[notification.IntegrationID]

		ig.DecryptSensitive(ng.Crypter)

		// dispatch entity that will actually send notification with given config
		n, err := notifier.FromIntegration(ig)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "can't resolve integration executor: %v", err)
		}

		err = n.Notify(ctx, notification, msg.Event)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't notify %s: %w", notification.ID, err))
		}

		metrics.MessagesSentCount.With(prometheus.Labels{
			metrics.IsSuccessful: strconv.FormatBool(err == nil),
			metrics.Type:         notification.IntegrationType,
		}).Inc()
	}

	if len(errs) > 0 {
		msg := errcommon.CollectErrors("NotifierGeneric.Notify", errs).Error()
		return nil, errcommon.StatusWithReason(codes.Internal, NotificationFailed, msg).Err()
	}

	return &emptypb.Empty{}, nil
}

func (ng *NotifierGeneric) validateNotifyReq(req *api.NotifyReq) (string, bool) {
	if len(req.GetNotifications()) == 0 {
		return "empty notifications", false
	}
	for i, n := range req.GetNotifications() {
		if n.GetEvent() == nil {
			return fmt.Sprintf("notifications[%d]'s event is nil", i), false
		}
	}
	return "", true
}
