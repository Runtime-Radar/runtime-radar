package updater

import (
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model"
)

type Dummy struct{}

func (*Dummy) Config() *model.Config {
	return &model.Config{}
}

func (*Dummy) SetConfig(_ *model.Config) {}

func (*Dummy) Equal(_, _ *model.Config) bool {
	return true
}

func (d *Dummy) SetOnUpdateFunc(_ func() error) {}
