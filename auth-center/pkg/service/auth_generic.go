package service

import (
	"context"
	"errors"
	"time"

	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/database"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/tokens"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type AuthGeneric struct {
	api.UnimplementedAuthControllerServer
	UserRepository   database.UserRepository
	TokenKey         []byte
	PasswordCheckMap map[string]struct{}
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

func (ag *AuthGeneric) SignIn(ctx context.Context, req *api.SignInReq) (resp *api.SignInResp, err error) {
	user, err := ag.UserRepository.GetByUsername(ctx, req.Username)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, status.Error(codes.NotFound, "user does not exist")
	case err != nil:
		return nil, status.Error(codes.Internal, "can't process user")
	}

	if user.AuthType == model.AuthTypeInternal {
		if verifyPassword(req.Password, user.HashedPassword) {
			tokenPair, err := tokens.GenerateTokenPair(*user, ag.TokenKey, ag.AccessTokenTTL, ag.RefreshTokenTTL)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "can't generate token pair")
			}

			return &api.SignInResp{
				AccessToken:  tokenPair.AccessTokenHash,
				RefreshToken: tokenPair.RefreshTokenHash,
				TokenType:    "Bearer",
			}, nil
		}
	}

	return nil, errcommon.StatusWithReason(codes.PermissionDenied, "UNCONFIRMED_USER", "unconfirmed user").Err()
}

func (ag *AuthGeneric) RefreshTokens(ctx context.Context, _ *emptypb.Empty) (resp *api.SignInResp, err error) {
	tData, err := tokens.RefreshTokenFromContext(ctx, ag.TokenKey)
	if err != nil {
		return nil, status.Error(codes.Internal, "can't get token data")
	}

	if tData.TokenType != jwt.TokenTypeRefresh || tData.RefreshTokenSalt != tokens.RefreshTokenSalt {
		return nil, errcommon.StatusWithReason(codes.Unauthenticated, errcommon.Unauthenticated, "unauthenticated").Err()
	}

	user, err := ag.UserRepository.GetByUsername(ctx, tData.Username)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errcommon.StatusWithReason(codes.Unauthenticated, errcommon.Unauthenticated, "user not found").Err()
	case err != nil:
		return nil, status.Error(codes.Internal, "can't process user")
	}

	if user.LastPasswordChangedAt.Unix() != tData.LastPasswordChangedAt {
		return nil, errcommon.StatusWithReason(codes.Unauthenticated, "ACCESS_AND_REFRESH_TOKENS_CHANGED", "access and refresh tokens was changed").Err()
	}

	tokenPair, err := tokens.GenerateTokenPair(*user, ag.TokenKey, ag.AccessTokenTTL, ag.RefreshTokenTTL)
	if err != nil {
		return nil, status.Error(codes.Internal, "can't generate token pair")
	}

	return &api.SignInResp{
		AccessToken:  tokenPair.AccessTokenHash,
		RefreshToken: tokenPair.RefreshTokenHash,
		TokenType:    "Bearer",
	}, nil
}
