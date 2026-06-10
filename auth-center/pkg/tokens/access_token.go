package tokens

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"google.golang.org/grpc/metadata"
)

type AccessToken struct {
	BaseToken
	Role  model.Role `json:"role"`
	Email string     `json:"email"`
}

func NewAccessToken(user model.User, ttl time.Duration) *AccessToken {
	nowTime := time.Now()
	return &AccessToken{
		BaseToken: BaseToken{
			UserID:                user.ID.String(),
			Username:              user.Username,
			TokenType:             jwt.TokenTypeAccess,
			AuthType:              user.AuthType.String(),
			LastPasswordChangedAt: user.LastPasswordChangedAt.Unix(),
			RegisteredClaims: jwtv5.RegisteredClaims{
				ExpiresAt: &jwtv5.NumericDate{Time: nowTime.Add(ttl)},
				IssuedAt:  &jwtv5.NumericDate{Time: nowTime},
			},
		},
		Role:  user.Role,
		Email: user.Email,
	}
}

func AccessTokenFromContext(ctx context.Context, key []byte) (*AccessToken, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	auth := md.Get(AuthorizationKey)
	if !ok || len(auth) == 0 {
		return nil, errors.New("no authorization token provided")
	}

	tokenStr, ok := strings.CutPrefix(auth[0], "Bearer ")
	if !ok {
		return nil, errors.New("unknown authorization format")
	}

	token, err := jwtv5.ParseWithClaims(tokenStr, &AccessToken{}, func(_ *jwtv5.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token was incorrectly parsed: %w", err)
	}

	t, ok := token.Claims.(*AccessToken)
	if !ok {
		return nil, errors.New("can't get token from claims")
	}

	return t, nil
}
