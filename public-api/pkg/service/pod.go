package service

import "github.com/runtime-radar/runtime-radar/kube-manager/api"

type Pod interface {
	api.PodControllerClient
}
