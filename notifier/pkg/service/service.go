package service

import (
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"google.golang.org/protobuf/proto"
)

const (
	defaultOrder = "created_at desc"
)

// gRPC errdetails.ErrorInfo.Reason codes used in service responses.
const (
	NameMustBeUnique        = "NAME_MUST_BE_UNIQUE"
	IntegrationInaccessible = "INTEGRATION_INACCESSIBLE"
	NotificationInUse       = "NOTIFICATION_IN_USE"
	NotificationFailed      = "NOTIFICATION_FAILED"
)

func hidePassword(req *api.Integration) *api.Integration {
	if req == nil {
		return nil
	}

	clone, ok := proto.Clone(req).(*api.Integration)
	if !ok {
		return req
	}

	const mask = "********"
	if email := clone.GetEmail(); email != nil {
		email.Password = mask
	}

	if webhook := clone.GetWebhook(); webhook != nil {
		webhook.Password = mask
	}

	return clone
}
