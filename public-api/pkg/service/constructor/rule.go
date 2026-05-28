package constructor

import (
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	enf_api "github.com/runtime-radar/runtime-radar/policy-enforcer/api"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/server/handler"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// unmarshalOptions returns a protojson.UnmarshalOptions configuration
// necessary for correctly parsing protojson and ignoring unknown fields.
// It enables the DiscardUnknown option so that missing fields are simply
// ignored during unmarshaling, ensuring that the parsing process does not
// fail due to unexpected fields.
func unmarshalOptions() protojson.UnmarshalOptions {
	return protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
}

func RuleCreate(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			st := status.Newf(codes.InvalidArgument, "can't read request body: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		req := &enf_api.Rule{}
		if err = unmarshalOptions().Unmarshal(body, req); err != nil {
			status := status.New(codes.InvalidArgument, err.Error())
			handler.StatusJSONResp(w, status)
			return
		}

		resp, err := svc.Create(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't create rule: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func RuleRead(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(mux.Vars(r)["id"])
		if err != nil {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, "id must be valid UUID"))
			return
		}

		resp, err := svc.Read(r.Context(), &enf_api.ReadRuleReq{Id: id.String()})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't read rule: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func RuleListPage(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &enf_api.ListRulePageReq{}

		pageNum, err := strconv.ParseUint(mux.Vars(r)["page_num"], 0, 32)
		if err != nil {
			status := status.Newf(codes.InvalidArgument, "invalid page_num: %v", err)
			handler.StatusJSONResp(w, status)
			return
		}
		req.PageNum = uint32(pageNum)

		q := r.URL.Query()

		if param := q.Get("page_size"); param != "" {
			pageSize, err := strconv.ParseUint(param, 0, 32)
			if err != nil {
				status := status.Newf(codes.InvalidArgument, "invalid page_size: %v", err)
				handler.StatusJSONResp(w, status)
				return
			}

			if pageSize == 0 {
				status := status.New(codes.InvalidArgument, "page_size must be more than 0")
				handler.StatusJSONResp(w, status)
				return
			}

			req.PageSize = uint32(pageSize)
		}

		req.Order = q.Get("order")

		resp, err := svc.ListPage(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't list rule page: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func RuleUpdate(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(mux.Vars(r)["id"])
		if err != nil {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, "id must be valid UUID"))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			st := status.Newf(codes.InvalidArgument, "can't read request body: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		req := &enf_api.Rule{}
		if err = unmarshalOptions().Unmarshal(body, req); err != nil {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, err.Error()))
			return
		}

		req.Id = id.String()

		resp, err := svc.Update(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't update rule: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func RuleDelete(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(mux.Vars(r)["id"])
		if err != nil {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, "id must be valid UUID"))
			return
		}

		resp, err := svc.Delete(r.Context(), &enf_api.DeleteRuleReq{Id: id.String()})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't delete rule: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func RuleNotifyTargetsInUse(svc service.Rule) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &enf_api.NotifyTargetsInUseReq{}

		targets := r.URL.Query()["targets"]
		if len(targets) == 0 {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, "targets is empty"))
			return
		}
		req.Targets = append(req.Targets, targets...)

		resp, err := svc.NotifyTargetsInUse(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't get notify target in use: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}
