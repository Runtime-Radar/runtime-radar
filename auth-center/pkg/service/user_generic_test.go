package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/tokens"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var securityEngineerRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000002")

// ctxWithCallerRole builds an incoming gRPC context carrying an access token issued for a user
// holding roleID, the way the interceptor would populate it for a real call.
func ctxWithCallerRole(t *testing.T, key []byte, roleID uuid.UUID) context.Context {
	t.Helper()

	user := model.User{
		Base:                  model.Base{ID: uuid.New()},
		Username:              "caller",
		RoleID:                roleID,
		Role:                  model.Role{ID: roleID},
		LastPasswordChangedAt: time.Now(),
	}

	pair, err := tokens.GenerateTokenPair(user, key, time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("can't generate token pair: %v", err)
	}

	md := metadata.Pairs(tokens.AuthorizationKey, "Bearer "+pair.AccessTokenHash)

	return metadata.NewIncomingContext(context.Background(), md)
}

func TestVerifyRoleAssignment(t *testing.T) {
	key := []byte("test-token-key")
	ug := &UserGeneric{TokenKey: key}

	tests := []struct {
		name          string
		callerRoleID  uuid.UUID
		currentRoleID uuid.UUID
		newRoleID     uuid.UUID
		wantReason    string
	}{
		{
			name:          "admin grants admin",
			callerRoleID:  model.AdminRoleID,
			currentRoleID: securityEngineerRoleID,
			newRoleID:     model.AdminRoleID,
		},
		{
			name:          "admin demotes admin",
			callerRoleID:  model.AdminRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     securityEngineerRoleID,
		},
		{
			name:          "non-admin keeps role unchanged",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: securityEngineerRoleID,
			newRoleID:     securityEngineerRoleID,
		},
		{
			name:          "non-admin creates user with own role",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: uuid.Nil,
			newRoleID:     securityEngineerRoleID,
		},
		{
			name:          "non-admin escalates a user to admin",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: securityEngineerRoleID,
			newRoleID:     model.AdminRoleID,
			wantReason:    RoleAssignmentRestricted,
		},
		{
			name:          "non-admin creates an admin",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: uuid.Nil,
			newRoleID:     model.AdminRoleID,
			wantReason:    RoleAssignmentRestricted,
		},
		{
			name:          "non-admin modifies an admin account",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     model.AdminRoleID,
			wantReason:    RoleAssignmentRestricted,
		},
		{
			name:          "non-admin demotes an admin",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     securityEngineerRoleID,
			wantReason:    RoleAssignmentRestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctxWithCallerRole(t, key, tt.callerRoleID)

			err := ug.verifyRoleAssignment(ctx, tt.currentRoleID, tt.newRoleID)

			if tt.wantReason == "" {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				return
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("Expected a gRPC status error, got %v", err)
			}
			if st.Code() != codes.PermissionDenied {
				t.Fatalf("Expected code %v, got %v", codes.PermissionDenied, st.Code())
			}
			if reason, _ := errcommon.ReasonFromStatus(st); reason != tt.wantReason {
				t.Fatalf("Expected reason %q, got %q", tt.wantReason, reason)
			}
		})
	}
}

func TestVerifyRoleAssignmentWithoutToken(t *testing.T) {
	ug := &UserGeneric{TokenKey: []byte("test-token-key")}

	err := ug.verifyRoleAssignment(context.Background(), uuid.Nil, model.AdminRoleID)

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected a gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("Expected code %v, got %v", codes.Unauthenticated, st.Code())
	}
}
