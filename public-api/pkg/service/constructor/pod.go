package constructor

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/server/handler"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PodGet(svc service.Pod) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		req := &api.GetPodReq{
			Name:      q.Get("name"),
			Namespace: q.Get("namespace"),
		}

		resp, err := svc.Get(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't get pod: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func PodListMeta(svc service.Pod) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := parseListPodMetaReq(r)

		resp, err := svc.ListMeta(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't list pod meta: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func PodListPage(svc service.Pod) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := parseListPodPageReq(r)
		if err != nil {
			handler.StatusJSONResp(w, status.Convert(err))
			return
		}

		resp, err := svc.ListPage(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't list pod page: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func parseListPodMetaReq(r *http.Request) *api.ListPodMetaReq {
	q := r.URL.Query()
	req := &api.ListPodMetaReq{}

	if field := q.Get("sort.field"); field != "" {
		req.Sort = &api.Sort{Field: field, Key: q.Get("sort.key")}
	}
	req.Namespaces = q["namespaces"]
	req.Nodes = q["nodes"]
	req.Pods = q["pods"]
	req.Containers = q["containers"]

	return req
}

func parseListPodPageReq(r *http.Request) (*api.ListPodPageReq, error) {
	pageNum, err := strconv.ParseUint(mux.Vars(r)["page_num"], 0, 32)
	if err != nil {
		return nil, status.Newf(codes.InvalidArgument, "invalid page_num: %v", err).Err()
	}

	q := r.URL.Query()
	req := &api.ListPodPageReq{
		PageNum: uint32(pageNum),
	}

	if field := q.Get("sort.field"); field != "" {
		req.Sort = &api.Sort{Field: field, Key: q.Get("sort.key")}
	}
	req.Namespaces = q["namespaces"]
	req.Nodes = q["nodes"]
	req.Pods = q["pods"]
	req.Containers = q["containers"]

	if param := q.Get("page_size"); param != "" {
		pageSize, err := strconv.ParseUint(param, 0, 32)
		if err != nil {
			return nil, status.Newf(codes.InvalidArgument, "invalid page_size: %v", err).Err()
		}
		if pageSize == 0 {
			return nil, status.New(codes.InvalidArgument, "page_size must be more than 0").Err()
		}
		req.PageSize = uint32(pageSize)
	}

	return req, nil
}
