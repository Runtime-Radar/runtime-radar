package service

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/auth-center/api"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/database"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/model"
	"github.com/runtime-radar/runtime-radar/auth-center/pkg/tokens"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type UserGeneric struct {
	api.UnimplementedUserControllerServer

	UserRepository   database.UserRepository
	TokenKey         []byte
	PasswordCheckMap map[string]struct{}
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

func (ug *UserGeneric) Read(ctx context.Context, userReq *api.ReadUserReq) (resp *api.UserResp, err error) {
	id, err := uuid.Parse(userReq.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse id: %v", err)
	}

	user, err := ug.UserRepository.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user does not exist")
		}
		return nil, status.Errorf(codes.Internal, "can't get user: %v", err)
	}

	resp = &api.UserResp{
		Id:                    user.ID.String(),
		Username:              user.Username,
		AuthType:              user.AuthType.String(),
		LastPasswordChangedAt: user.LastPasswordChangedAt.Unix(),
		Email:                 user.Email,
		RoleId:                user.RoleID.String(),
		Role:                  fillRoleResp(&user.Role),
	}

	return resp, nil
}

func (ug *UserGeneric) ReadList(ctx context.Context, _ *emptypb.Empty) (resp *api.UserListResp, err error) {
	resp = &api.UserListResp{}

	users, err := ug.UserRepository.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get users: %v", err)
	}

	for _, user := range users {
		resp.Users = append(resp.Users, &api.UserResp{
			Id:                    user.ID.String(),
			Username:              user.Username,
			AuthType:              user.AuthType.String(),
			LastPasswordChangedAt: user.LastPasswordChangedAt.Unix(),
			Email:                 user.Email,
			RoleId:                user.RoleID.String(),
			Role:                  fillRoleResp(&user.Role),
		})
	}

	return resp, nil
}

func (ug *UserGeneric) Create(ctx context.Context, req *api.CreateUserReq) (resp *api.UserResp, err error) {
	authType, isValid := model.ValidateAuthType(req.AuthType)
	if !isValid {
		return nil, status.Errorf(codes.InvalidArgument, "invalid auth type: %v", req.AuthType)
	}

	var id uuid.UUID
	if req.GetId() != "" {
		id, err = uuid.Parse(req.GetId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "can't parse id: %v", err)
		}
	}

	if req.RoleId == "" {
		return nil, status.Error(codes.FailedPrecondition, "role id is empty")
	}
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse role id: %v", err)
	}

	if err := ug.verifyUserCreation(ctx); err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse email: %v", err)
	}

	reason := newPasswordCheck(req.Password, ug.PasswordCheckMap)
	if reason != "" {
		return nil, errcommon.StatusWithReason(codes.Aborted, reason, "invalid password").Err()
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "can't hash password")
	}

	userToCreate := &model.User{
		Base:                  model.Base{ID: id},
		Username:              req.Username,
		Email:                 req.Email,
		AuthType:              authType,
		HashedPassword:        hashedPassword,
		RoleID:                roleID,
		LastPasswordChangedAt: time.Now(),
	}

	if req.MappingRoleId != "" {
		id, err = uuid.Parse(req.MappingRoleId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "can't parse mapping role id")
		}
		userToCreate.MappingRoleID = &id
	}

	err = ug.UserRepository.Create(ctx, userToCreate)
	switch {
	case errors.Is(err, database.ErrUsernameAlreadyExists):
		return nil, status.Error(codes.InvalidArgument, "username already exists")
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return nil, status.Error(codes.AlreadyExists, "user with ID already exists")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return nil, status.Error(codes.Internal, "role does not exist")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "user creating failed: %v", err)
	}

	user, err := ug.UserRepository.GetByID(ctx, userToCreate.ID) // Do this to get nested struct Role
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "can't get created user: %v", req.Id)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create user: %v", err)
	}

	resp = &api.UserResp{
		Id:                    user.ID.String(),
		Username:              user.Username,
		AuthType:              authType.String(),
		LastPasswordChangedAt: userToCreate.LastPasswordChangedAt.Unix(),
		Email:                 user.Email,
		RoleId:                user.RoleID.String(),
		Role:                  fillRoleResp(&user.Role),
	}

	return resp, nil
}

// verifyUserCreation checks that the caller may add an account. Only an administrator creates users,
// since creating one also assigns its role and would otherwise hand out administrator itself.
func (ug *UserGeneric) verifyUserCreation(ctx context.Context) error {
	token, err := tokens.AccessTokenFromContext(ctx, ug.TokenKey)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "can't get token: %v", err)
	}

	if token.Role.ID != model.AdminRoleID {
		return errcommon.StatusWithReason(codes.PermissionDenied, UserManagementRestricted,
			"can't create users").Err()
	}

	return nil
}

// verifyUserUpdate checks that the caller may apply the requested change to target. Only an
// administrator edits other accounts and hands out roles, otherwise users:update escalates.
func (ug *UserGeneric) verifyUserUpdate(ctx context.Context, target *model.User, newRoleID uuid.UUID) error {
	token, err := tokens.AccessTokenFromContext(ctx, ug.TokenKey)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "can't get token: %v", err)
	}

	if token.Role.ID == model.AdminRoleID {
		if target.RoleID == model.AdminRoleID && newRoleID != model.AdminRoleID {
			return ug.verifyNotLastAdmin(ctx)
		}

		return nil
	}

	if token.UserID != target.ID.String() {
		return errcommon.StatusWithReason(codes.PermissionDenied, UserManagementRestricted,
			"can't modify another user").Err()
	}

	if newRoleID != target.RoleID {
		return errcommon.StatusWithReason(codes.PermissionDenied, RoleAssignmentRestricted,
			"can't change your own role").Err()
	}

	return nil
}

// verifyNotLastAdmin rejects demoting the only administrator left, the same way Delete guards the last one.
func (ug *UserGeneric) verifyNotLastAdmin(ctx context.Context) error {
	adminUsers, err := ug.UserRepository.GetUsersByRoleID(ctx, model.AdminRoleID)
	if err != nil {
		return status.Error(codes.Internal, "internal error")
	}

	if len(adminUsers) == 1 {
		return errcommon.StatusWithReason(codes.PermissionDenied, LastAdminRemovingDenied,
			"can't demote last administrator").Err()
	}

	return nil
}

func (ug *UserGeneric) Update(ctx context.Context, req *api.UpdateUserReq) (resp *api.UserResp, err error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse id: %v", err)
	}

	target, err := ug.UserRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user does not exist")
		}
		return nil, status.Errorf(codes.Internal, "can't get user: %v", err)
	}

	_, err = mail.ParseAddress(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse email: %v", err)
	}

	if req.RoleId == "" {
		return nil, status.Error(codes.FailedPrecondition, "role id is empty")
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse role id")
	}

	if err := ug.verifyUserUpdate(ctx, target, roleID); err != nil {
		return nil, err
	}

	user := &model.User{
		Base:   model.Base{ID: id},
		Email:  req.Email,
		RoleID: roleID,
	}

	err = ug.UserRepository.Update(ctx, user)
	switch {
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return nil, status.Error(codes.Internal, "role does not exist")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "can't update user: %v", err)
	}

	user, err = ug.UserRepository.GetByID(ctx, user.ID) // Do this to get nested struct Role and another field
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get updated user: %v", err)
	}

	resp = &api.UserResp{
		Id:                    user.ID.String(),
		Username:              user.Username,
		AuthType:              user.AuthType.String(),
		LastPasswordChangedAt: user.LastPasswordChangedAt.Unix(),
		Email:                 user.Email,
		RoleId:                user.RoleID.String(),
		Role:                  fillRoleResp(&user.Role),
	}

	return resp, nil
}

func (ug *UserGeneric) Delete(ctx context.Context, req *api.DeleteUserReq) (resp *api.DeleteUserResp, err error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse user id: %v", err)
	}

	user, err := ug.UserRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user does not exist")
		}
		return nil, status.Errorf(codes.Internal, "can't get user: %v", err)
	}

	if user.RoleID.String() == model.AdminRoleIDStr {
		adminUsers, err := ug.UserRepository.GetUsersByRoleID(ctx, user.RoleID)
		if err != nil {
			return nil, status.Error(codes.Internal, "internal error")
		}
		if len(adminUsers) == 1 {
			return nil, errcommon.StatusWithReason(codes.PermissionDenied, LastAdminRemovingDenied, "can't delete last administrator").Err()
		}
	}

	err = ug.UserRepository.Delete(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't delete user: %v", err)
	}

	return &api.DeleteUserResp{Id: id.String()}, nil
}

func (ug *UserGeneric) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (resp *api.SignInResp, err error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse user id: %v", err)
	}

	user, err := ug.UserRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "user does not exist")
		}
		return nil, status.Errorf(codes.Internal, "can't get user: %v", err)
	}

	token, err := tokens.AccessTokenFromContext(ctx, ug.TokenKey)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "can't get token: %v", err)
	}
	if token.UserID != user.ID.String() && token.Role.ID != model.AdminInitID {
		return nil, errcommon.StatusWithReason(codes.PermissionDenied, "THIRD_PARTY_PASSWORD_CHANGE_IS_RESTRICTED", "can't change third party password").Err()
	}

	if user.AuthType == model.AuthTypeLDAP {
		return nil, errcommon.StatusWithReason(codes.Aborted, "LDAP_TYPE_USER_PASSWORD_CHANGE_IS_RESTRICTED", "can't change password for LDAP users").Err()
	}

	if verifyPassword(req.Password, user.HashedPassword) {
		return nil, errcommon.StatusWithReason(codes.Aborted, "PASSWORD_HAS_BEEN_USED_BEFORE", "the old and new passwords are equal").Err()
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "can't hash password")
	}

	reason := newPasswordCheck(req.Password, ug.PasswordCheckMap)
	if reason != "" {
		return nil, errcommon.StatusWithReason(codes.Aborted, reason, "invalid password").Err()
	}

	err = ug.UserRepository.ChangePassword(ctx, hashedPassword, id)
	if err != nil {
		return nil, errcommon.StatusWithReason(codes.Internal, "PASSWORD_UPDATE_FAILED", "can't change password").Err()
	}

	user, err = ug.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get user: %v", err)
	}

	tokenPair, err := tokens.GenerateTokenPair(*user, ug.TokenKey, ug.AccessTokenTTL, ug.RefreshTokenTTL)
	if err != nil {
		return nil, status.Error(codes.Internal, "can't generate token pair")
	}

	return &api.SignInResp{
		AccessToken:  tokenPair.AccessTokenHash,
		RefreshToken: tokenPair.RefreshTokenHash,
		TokenType:    "Bearer",
	}, nil
}
