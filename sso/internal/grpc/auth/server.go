package authgrpc

import (
	"context"
	"errors"
	domain "sso/internal/domain"
	customErr "sso/internal/domain/errors"

	ssov1 "github.com/sj-shoff/sso_proto/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth auth
}

func Register(gRPCServer *grpc.Server, auth auth) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, in *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if in.GetUsername() == "" || in.GetPassword() == "" || in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "username, password and app_id required")
	}

	_, token, _, err := s.auth.Login(ctx, in.GetUsername(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, customErr.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "login failed")
	}
	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if in.GetUsername() == "" || in.GetPassword() == "" || in.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "username, password and app_id required")
	}

	roleStr := in.GetRole()
	if roleStr == "" {
		roleStr = string(domain.RoleViewer)
	}
	if !domain.IsValidRole(roleStr) {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetUsername(), in.GetPassword(), domain.UserRole(roleStr), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, customErr.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "register failed")
	}
	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) GetUsers(ctx context.Context, in *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	users, err := s.auth.GetUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get users")
	}

	respUsers := make([]*ssov1.User, len(users))
	for i, u := range users {
		respUsers[i] = &ssov1.User{
			Id:       u.ID,
			Username: u.Username,
			Role:     string(u.Role),
		}
	}
	return &ssov1.GetUsersResponse{Users: respUsers}, nil
}

func (s *serverAPI) UpdateUserRole(ctx context.Context, in *ssov1.UpdateRoleRequest) (*ssov1.UpdateRoleResponse, error) {
	if in.GetUserId() == 0 || in.GetNewRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and new_role required")
	}
	if !domain.IsValidRole(in.GetNewRole()) {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	err := s.auth.UpdateUserRole(ctx, in.GetUserId(), domain.UserRole(in.GetNewRole()))
	if err != nil {
		if errors.Is(err, customErr.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "update role failed")
	}
	return &ssov1.UpdateRoleResponse{}, nil
}
