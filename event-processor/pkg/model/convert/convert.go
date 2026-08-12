package convert

import (
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	"github.com/runtime-radar/runtime-radar/event-processor/pkg/model"
)

func DetectorToProto(d *model.Detector) *api.Detector {
	det := &api.Detector{
		Id:             d.ID,
		Name:           d.Name,
		Description:    d.Description,
		Version:        uint32(d.Version),
		Author:         d.Author,
		Contact:        d.Contact,
		License:        d.License,
		TacticsCovered: MitreTacticsToProto(d.TacticsCovered),
	}

	return det
}

func DetectorsToProto(ds []*model.Detector) (ps []*api.Detector) {
	for _, d := range ds {
		ps = append(ps, DetectorToProto(d))
	}

	return
}

func MitreTacticsToProto(mts model.MitreTactics) []*api.MitreTactic {
	res := make([]*api.MitreTactic, 0, len(mts))

	for _, mt := range mts {
		apiMT := &api.MitreTactic{
			Id:         mt.GetId(),
			Techniques: make([]string, 0, len(mt.GetTechniques())),
		}

		for _, t := range mt.GetTechniques() {
			apiMT.Techniques = append(apiMT.Techniques, t)
		}

		res = append(res, apiMT)
	}

	return res
}
