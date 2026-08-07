package service

import (
	"context"
	"errors"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/model/convert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
)

type NodeGeneric struct {
	api.UnimplementedNodeControllerServer

	Nodes informers.Getter[*v1.Node]
}

func (ng *NodeGeneric) Get(_ context.Context, req *api.GetNodeReq) (resp *api.GetNodeResp, err error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "node name is empty")
	}

	node, err := ng.Nodes.Get("", req.GetName())
	switch {
	case errors.Is(err, informers.ErrNotFound):
		return nil, status.Error(codes.NotFound, "node not found")
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}

	apiNode, err := convert.NodeToAPI(node)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.GetNodeResp{
		Node: apiNode,
	}, nil
}

func (ng *NodeGeneric) ListMeta(_ context.Context, req *api.ListNodeMetaReq) (*api.ListNodeMetaResp, error) {
	opts, _, err := validateParams(nil, req.Names, nil, nil, req.GetSort())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	list, total := ng.Nodes.List()
	prepared, err := ng.withOpts(list, opts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var nodes []*api.ListNodeMetaResp_Node
	for _, node := range prepared {
		nodes = append(nodes, &api.ListNodeMetaResp_Node{
			Name:     node.GetName(),
			Ip:       ng.getNodeAddress(node, v1.NodeInternalIP),
			Hostname: ng.getNodeAddress(node, v1.NodeHostName),
			Uid:      string(node.UID),
		})
	}

	return &api.ListNodeMetaResp{
		Total: uint32(total),
		Nodes: nodes,
	}, nil
}

func (ng *NodeGeneric) ListPage(_ context.Context, req *api.ListNodePageReq) (*api.ListNodePageResp, error) {
	opts, _, err := validateParams(nil, req.Names, nil, nil, req.GetSort())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	opts.PageSize = req.PageSize
	opts.PageNum = req.PageNum

	list, total := ng.Nodes.List()
	nodes, err := ng.withOpts(list, opts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	apiNodes, err := convert.NodesToAPI(nodes)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ListNodePageResp{
		Total: uint32(total),
		Nodes: apiNodes,
	}, nil
}

func (ng *NodeGeneric) getNodeAddress(node *v1.Node, addressType v1.NodeAddressType) string {
	if node == nil {
		return ""
	}

	for _, a := range node.Status.Addresses {
		if a.Type == addressType {
			return a.Address
		}
	}

	return ""
}

func (ng *NodeGeneric) withOpts(nodes []*v1.Node, opts *ListOpts) ([]*v1.Node, error) {
	var filtered []*v1.Node
	for _, node := range nodes {
		isMatched, err := opts.isNodeMatched(node)
		if err != nil {
			return nil, err
		}

		if !isMatched {
			continue
		}

		filtered = append(filtered, node)
	}

	return withPagination(withSort(filtered, opts), opts), nil
}
