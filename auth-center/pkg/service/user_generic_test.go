package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/database"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/tokens"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	securityEngineerRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	cicdRoleID             = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

// stubUserRepository answers with a fixed number of administrators; the other methods are unused here.
type stubUserRepository struct {
	database.UserRepository

	admins int
}

func (s *stubUserRepository) GetUsersByRoleID(_ context.Context, roleID uuid.UUID) ([]*model.User, error) {
	users := make([]*model.User, 0, s.admins)
	for i := 0; i < s.admins; i++ {
		users = append(users, &model.User{RoleID: roleID, Role: model.Role{ID: roleID}})
	}

	return users, nil
}

// ctxWithCaller builds an incoming gRPC context carrying an access token issued for the given user,
// the way the interceptor would populate it for a real call.
func ctxWithCaller(t *testing.T, key []byte, userID, roleID uuid.UUID) context.Context {
	t.Helper()

	user := model.User{
		Base:                  model.Base{ID: userID},
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

func TestVerifyUserModification(t *testing.T) {
	key := []byte("test-token-key")

	tests := []struct {
		name         string
		callerRoleID uuid.UUID
		// create leaves the request without an existing user; targetIsCaller tells whether the
		// updated account is the caller's own one.
		create         bool
		targetIsCaller bool
		currentRoleID  uuid.UUID
		newRoleID      uuid.UUID
		admins         int
		wantReason     string
	}{
		{
			name:         "admin creates an admin",
			callerRoleID: model.AdminRoleID,
			create:       true,
			newRoleID:    model.AdminRoleID,
		},
		{
			name:          "admin promotes another user",
			callerRoleID:  model.AdminRoleID,
			currentRoleID: securityEngineerRoleID,
			newRoleID:     model.AdminRoleID,
		},
		{
			name:          "admin demotes another admin",
			callerRoleID:  model.AdminRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     securityEngineerRoleID,
			admins:        2,
		},
		{
			name:          "admin demotes the last admin",
			callerRoleID:  model.AdminRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     securityEngineerRoleID,
			admins:        1,
			wantReason:    LastAdminRemovingDenied,
		},
		{
			name:           "non-admin edits its own account",
			callerRoleID:   securityEngineerRoleID,
			targetIsCaller: true,
			currentRoleID:  securityEngineerRoleID,
			newRoleID:      securityEngineerRoleID,
		},
		{
			name:         "non-admin creates a user",
			callerRoleID: securityEngineerRoleID,
			create:       true,
			newRoleID:    securityEngineerRoleID,
			wantReason:   UserManagementRestricted,
		},
		{
			name:          "non-admin edits another user",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: cicdRoleID,
			newRoleID:     cicdRoleID,
			wantReason:    UserManagementRestricted,
		},
		{
			name:          "non-admin edits an admin account",
			callerRoleID:  securityEngineerRoleID,
			currentRoleID: model.AdminRoleID,
			newRoleID:     model.AdminRoleID,
			wantReason:    UserManagementRestricted,
		},
		{
			name:           "non-admin escalates itself to admin",
			callerRoleID:   securityEngineerRoleID,
			targetIsCaller: true,
			currentRoleID:  securityEngineerRoleID,
			newRoleID:      model.AdminRoleID,
			wantReason:     RoleAssignmentRestricted,
		},
		{
			name:           "non-admin changes its own role",
			callerRoleID:   securityEngineerRoleID,
			targetIsCaller: true,
			currentRoleID:  securityEngineerRoleID,
			newRoleID:      cicdRoleID,
			wantReason:     RoleAssignmentRestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callerID := uuid.New()
			ctx := ctxWithCaller(t, key, callerID, tt.callerRoleID)
			ug := &UserGeneric{TokenKey: key, UserRepository: &stubUserRepository{admins: tt.admins}}

			var current *model.User
			if !tt.create {
				targetID := uuid.New()
				if tt.targetIsCaller {
					targetID = callerID
				}
				current = &model.User{Base: model.Base{ID: targetID}, RoleID: tt.currentRoleID}
			}

			err := ug.verifyUserModification(ctx, current, tt.newRoleID)

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

func TestVerifyUserModificationWithoutToken(t *testing.T) {
	ug := &UserGeneric{TokenKey: []byte("test-token-key")}

	err := ug.verifyUserModification(context.Background(), nil, model.AdminRoleID)

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected a gRPC status error, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("Expected code %v, got %v", codes.Unauthenticated, st.Code())
	}
}
