package interceptor

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Recovery intercepts panic so that the app can operate normally after any occasional panic in handlers.
func Recovery(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = status.Errorf(codes.Internal, "panic caught while calling '%s': %v", info.FullMethod, p)
			log.Error().Str("stacktrace", string(debug.Stack())).Msgf("Panic caught while calling '%s': %v", info.FullMethod, p)

			if e := log.Debug(); e.Enabled() {
				debug.PrintStack()
			}
		}
	}()

	return handler(ctx, req)
}

const CorrelationHeader = "correlation-id"

func Correlation(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var corrID string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(CorrelationHeader)) == 0 {
		corrID = uuid.NewString()
	} else {
		corrID = md.Get(CorrelationHeader)[0]
	}

	ctx = metadata.AppendToOutgoingContext(ctx, CorrelationHeader, corrID)

	return handler(ctx, req)
}

func CorrelationIDFromContext(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok || len(md.Get(CorrelationHeader)) == 0 {
		return uuid.Nil, errors.New("no correlation id provided")
	}

	corrID, err := uuid.Parse(md.Get(CorrelationHeader)[0])
	if err != nil {
		return uuid.Nil, err
	}

	return corrID, nil
}
