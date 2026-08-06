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

func NodeGet(svc service.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &api.GetNodeReq{
			Name: r.URL.Query().Get("name"),
		}

		resp, err := svc.Get(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't get node: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func NodeListMeta(svc service.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := parseListNodeMetaReq(r)

		resp, err := svc.ListMeta(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't list node meta: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func NodeListPage(svc service.Node) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := parseListNodePageReq(r)
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
			st := status.Newf(codes.Internal, "can't list node page: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func parseListNodeMetaReq(r *http.Request) *api.ListNodeMetaReq {
	q := r.URL.Query()
	req := &api.ListNodeMetaReq{}

	if field := q.Get("sort.field"); field != "" {
		req.Sort = &api.Sort{Field: field, Key: q.Get("sort.key")}
	}
	req.Names = q["names"]

	return req
}

func parseListNodePageReq(r *http.Request) (*api.ListNodePageReq, error) {
	pageNum, err := strconv.ParseUint(mux.Vars(r)["page_num"], 0, 32)
	if err != nil {
		return nil, status.Newf(codes.InvalidArgument, "invalid page_num: %v", err).Err()
	}

	q := r.URL.Query()
	req := &api.ListNodePageReq{
		PageNum: uint32(pageNum),
	}

	if field := q.Get("sort.field"); field != "" {
		req.Sort = &api.Sort{Field: field, Key: q.Get("sort.key")}
	}
	req.Names = q["names"]

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
