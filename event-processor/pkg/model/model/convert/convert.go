package convert

import (
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
)

func DetectorToProto(d *model.Detector) *api.Detector {
	det := &api.Detector{
		Id:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Version:     uint32(d.Version),
		Author:      d.Author,
		Contact:     d.Contact,
		License:     d.License,
	}

	return det
}

func DetectorsToProto(ds []*model.Detector) (ps []*api.Detector) {
	for _, d := range ds {
		ps = append(ps, DetectorToProto(d))
	}

	return
}
