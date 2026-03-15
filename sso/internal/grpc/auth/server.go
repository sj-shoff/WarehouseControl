package authgrpc

import (
	"context"
	"errors"

	"sso/internal/domain"
	customErr "sso/internal/domain/errors"

	ssov1 "github.com/sj-shoff/sso_proto/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth authProvider
}

func Register(gRPCServer *grpc.Server, auth authProvider) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, in *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(in); err != nil {
		return nil, err
	}

	_, token, _, err := s.auth.Login(ctx, in.GetUsername(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, customErr.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, customErr.ErrInvalidCredentials.Error())
		}
		return nil, status.Error(codes.Internal, customErr.ErrInternal.Error())
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(in); err != nil {
		return nil, err
	}

	role := domain.UserRole(in.GetRole())
	if role == "" {
		role = domain.RoleViewer
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetUsername(), in.GetPassword(), role, int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, customErr.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, customErr.ErrUserExists.Error())
		}
		return nil, status.Error(codes.Internal, customErr.ErrInternal.Error())
	}

	return &ssov1.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) GetUsers(ctx context.Context, _ *ssov1.GetUsersRequest) (*ssov1.GetUsersResponse, error) {
	users, err := s.auth.GetUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, customErr.ErrInternal.Error())
	}

	respUsers := make([]*ssov1.User, 0, len(users))
	for _, u := range users {
		respUsers = append(respUsers, &ssov1.User{
			Id:       u.ID,
			Username: u.Username,
			Role:     string(u.Role),
		})
	}

	return &ssov1.GetUsersResponse{Users: respUsers}, nil
}

func (s *serverAPI) UpdateUserRole(ctx context.Context, in *ssov1.UpdateRoleRequest) (*ssov1.UpdateRoleResponse, error) {
	if in.GetUserId() <= 0 || in.GetNewRole() == "" {
		return nil, status.Error(codes.InvalidArgument, customErr.ErrInvalidInput.Error())
	}

	if !domain.IsValidRole(in.GetNewRole()) {
		return nil, status.Error(codes.InvalidArgument, customErr.ErrInvalidInput.Error())
	}

	err := s.auth.UpdateUserRole(ctx, in.GetUserId(), domain.UserRole(in.GetNewRole()))
	if err != nil {
		if errors.Is(err, customErr.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, customErr.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, customErr.ErrInternal.Error())
	}

	return &ssov1.UpdateRoleResponse{}, nil
}

func validateLogin(in *ssov1.LoginRequest) error {
	if in.GetUsername() == "" || in.GetPassword() == "" || in.GetAppId() <= 0 {
		return status.Error(codes.InvalidArgument, customErr.ErrInvalidInput.Error())
	}
	return nil
}

func validateRegister(in *ssov1.RegisterRequest) error {
	if in.GetUsername() == "" || in.GetPassword() == "" || in.GetAppId() <= 0 {
		return status.Error(codes.InvalidArgument, customErr.ErrInvalidInput.Error())
	}
	if in.GetRole() != "" && !domain.IsValidRole(in.GetRole()) {
		return status.Error(codes.InvalidArgument, customErr.ErrInvalidInput.Error())
	}
	return nil
}
