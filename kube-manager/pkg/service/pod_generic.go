package service

import (
	"context"
	"errors"
	"strings"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/client"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model/convert"
	"github.com/runtime-radar/runtime-radar/lib/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

type PodGeneric struct {
	api.UnimplementedPodControllerServer

	Pods       informers.Getter[*v1.Pod]
	Kubernetes *client.Kubernetes
}

func (pg *PodGeneric) Get(_ context.Context, req *api.GetPodReq) (resp *api.GetPodResp, err error) {
	if req.GetName() == "" || req.GetNamespace() == "" {
		return nil, status.Error(codes.InvalidArgument, "pod name or namespace is empty")
	}

	pod, err := pg.Pods.Get(req.GetNamespace(), req.GetName())
	switch {
	case errors.Is(err, informers.ErrNotFound):
		return nil, status.Error(codes.NotFound, "pod not found")
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	apiPod, err := convert.PodToAPI(pod)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.GetPodResp{
		Pod: apiPod,
	}, nil
}

func (pg *PodGeneric) ListMeta(_ context.Context, req *api.ListPodMetaReq) (*api.ListPodMetaResp, error) {
	opts, namespaces, err := validateParams(req.Namespaces, req.Nodes, req.Pods, req.Containers, req.GetSort())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	list, total := pg.Pods.List(namespaces...)

	prepared, err := pg.withOpts(list, opts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var pods []*api.ListPodMetaResp_Pod
	for _, pod := range prepared {
		containers := make([]*api.ListPodMetaResp_Pod_Container, 0, len(pod.Spec.Containers))
		for _, c := range pod.Spec.Containers {
			containers = append(containers, &api.ListPodMetaResp_Pod_Container{
				Name:        c.Name,
				ImageRef:    c.Image,
				ImageDigest: digest(pod, c.Name),
			})
		}

		pods = append(pods, &api.ListPodMetaResp_Pod{
			Name:       pod.GetName(),
			Namespace:  pod.GetNamespace(),
			Uid:        string(pod.UID),
			NodeName:   pod.Spec.NodeName,
			Phase:      string(pod.Status.Phase),
			Containers: containers,
		})
	}

	return &api.ListPodMetaResp{
		Total: uint32(total),
		Pods:  pods,
	}, nil
}

func (pg *PodGeneric) ListPage(_ context.Context, req *api.ListPodPageReq) (*api.ListPodPageResp, error) {
	opts, namespaces, err := validateParams(req.Namespaces, req.Nodes, req.Pods, req.Containers, req.GetSort())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	opts.PageSize = req.PageSize
	opts.PageNum = req.PageNum

	prepared, total := pg.Pods.List(namespaces...)

	pods, err := pg.withOpts(prepared, opts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	apiPods, err := convert.PodsToAPI(pods)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ListPodPageResp{
		Total: uint32(total),
		Pods:  apiPods,
	}, nil
}

func (pg *PodGeneric) Kill(ctx context.Context, req *api.KillPodReq) (*emptypb.Empty, error) {
	if req.GetName() == "" || req.GetNamespace() == "" {
		return nil, status.Error(codes.InvalidArgument, "both name and namespace should be present")
	}

	err := pg.Kubernetes.DeletePod(ctx, req.GetNamespace(), req.GetName(), true)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return nil, status.Errorf(codes.Internal, "can't kill pod: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (pg *PodGeneric) withOpts(pods []*v1.Pod, opts *ListOpts) ([]*v1.Pod, error) {
	var filtered []*v1.Pod
	for _, pod := range pods {
		isMatched, err := opts.isPodMatched(pod)
		if err != nil {
			return nil, err
		}

		if !isMatched {
			continue
		}

		filtered = append(filtered, pod)
	}

	return withPagination(withSort(filtered, opts), opts), nil
}

func digest(pod *v1.Pod, containerName string) string {
	for _, st := range pod.Status.ContainerStatuses {
		if st.Name == containerName {
			return extractDigest(st.ImageID)
		}
	}

	return ""
}

func extractDigest(imageID string) string {
	switch parts := strings.SplitN(imageID, "@", 2); {
	case len(parts) != 2:
		return ""

	case util.IsValidDigest(parts[1]):
		return parts[1]

	default:
		return ""
	}
}
