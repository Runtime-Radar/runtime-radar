package constructor

import (
	"io"
	"net/http"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/server/handler"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ConfigAdd(svc service.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			st := status.Newf(codes.InvalidArgument, "can't read request body: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		req := &api.Config{}
		if err = protojson.Unmarshal(body, req); err != nil {
			handler.StatusJSONResp(w, status.New(codes.InvalidArgument, err.Error()))
			return
		}

		resp, err := svc.Add(r.Context(), req)
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't add config: %v", err)
			handler.StatusJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}

func ConfigRead(svc service.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := svc.Read(r.Context(), &emptypb.Empty{})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				handler.StatusJSONResp(w, st)
				return
			}
			st := status.Newf(codes.Internal, "can't read config: %v", err)
			handler.ErrorJSONResp(w, st)
			return
		}

		handler.SendProtoResp(w, resp)
	})
}
