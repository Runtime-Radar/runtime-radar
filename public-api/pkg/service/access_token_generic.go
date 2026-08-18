package service

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/database"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model"
	"github.com/runtime-radar/runtime-radar/public-api/pkg/model/convert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	AccessTokenSizeBytes = 64
	// SaltSizeBytes is the allowed salt size in bytes, used to compute access token hash.
	SaltSizeBytes = 64
)

type AccessTokenGeneric struct {
	AccessTokenSalt []byte
	// JWTKey is used to parse JWTs in order to extract user ID.
	JWTKey []byte

	AccessTokenRepository database.AccessTokenRepository
}

func (ac *AccessTokenGeneric) Create(ctx context.Context, req *model.CreateAccessTokenReq) (uuid.UUID, string, error) {
	if reason, ok := ac.validateCreateReq(req); !ok {
		return uuid.Nil, "", status.Error(codes.InvalidArgument, reason)
	}

	token := hex.EncodeToString(security.Rand(AccessTokenSizeBytes))
	hashed := security.HashSaltedSHA512AsHex([]byte(token), ac.AccessTokenSalt)
	at := &model.AccessToken{
		model.Base{},
		req.Name,
		req.UserID,
		hashed,
		req.Permissions,
		req.ExpiresAt,
		nil,
	}

	if err := ac.AccessTokenRepository.Add(ctx, at); err != nil {
		if errors.Is(err, model.ErrTokenNameInUse) {
			return uuid.Nil, "", errcommon.StatusWithReason(codes.AlreadyExists, NameMustBeUnique, "token name must be unique").Err()
		}
		return uuid.Nil, "", status.Errorf(codes.Internal, "can't save token: %v", err)
	}

	return at.ID, token, nil
}

func (ac *AccessTokenGeneric) ListPage(ctx context.Context, pageNum, pageSize int, order string) ([]*model.AccessTokenResp, int, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, 0, status.Errorf(codes.Unauthenticated, "can't get user id: %v", err)
	}

	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	if pageNum <= 0 {
		pageNum = DefaultPageNum
	}

	filter := gorm.Expr("user_id = ?", userID)
	total, err := ac.AccessTokenRepository.GetCount(ctx, filter)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "can't get access token count: %v", err)
	}

	accessTokens, err := ac.AccessTokenRepository.GetPage(ctx, pageNum, pageSize, filter, order)
	if err != nil {
		if errors.Is(err, database.ErrInvalidOrder) {
			return nil, 0, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, 0, status.Errorf(codes.Internal, "can't get token list: %v", err)
	}

	return convert.AccessTokensToResponse(accessTokens), total, nil
}

func (ac *AccessTokenGeneric) GetByID(ctx context.Context, id uuid.UUID) (*model.AccessTokenResp, error) {
	if id == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "no id provided")
	}

	at, err := ac.AccessTokenRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get token: %v", err)
	}

	return &model.AccessTokenResp{
		at.ID,
		at.Name,
		at.UserID,
		at.Permissions,
		at.ExpiresAt,
		at.CreatedAt,
		at.InvalidatedAt,
	}, nil
}

func (ac *AccessTokenGeneric) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return status.Error(codes.InvalidArgument, "no id provided")
	}

	if err := ac.AccessTokenRepository.Delete(ctx, id); err != nil {
		return status.Errorf(codes.Internal, "can't delete token: %v", err)
	}

	return nil
}
func (ac *AccessTokenGeneric) InvalidateAll(ctx context.Context) error {
	return ac.AccessTokenRepository.InvalidateAll(ctx)
}

func (ac *AccessTokenGeneric) validateCreateReq(req *model.CreateAccessTokenReq) (reason string, ok bool) {
	if req.Name == "" {
		return "no name", false
	}
	if req.Permissions == nil {
		return "no permissions", false
	}
	if req.UserID == uuid.Nil {
		return "no user id", false
	}
	if req.ExpiresAt != nil {
		if req.ExpiresAt.IsZero() {
			return "no expires at", false
		}
		if req.ExpiresAt.Before(time.Now()) {
			return "expired token", false
		}
	}

	return "", true
}
