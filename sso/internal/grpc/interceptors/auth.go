package interceptors

import (
	"context"
	"fmt"
	customErr "sso/internal/domain/errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAdminInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		protectedMethods := map[string]bool{
			"/sso.Auth/Register":       true,
			"/sso.Auth/UpdateUserRole": true,
		}

		if protectedMethods[info.FullMethod] {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, customErr.ErrMetadataMiss.Error())
			}

			authHeader := md.Get("authorization")
			if len(authHeader) == 0 {
				return nil, status.Error(codes.Unauthenticated, customErr.ErrAuthToken.Error())
			}

			tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return nil, status.Error(codes.Unauthenticated, customErr.ErrInvalidToken.Error())
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, status.Error(codes.Internal, customErr.ErrGetClaims.Error())
			}

			role, ok := claims["role"].(string)
			if !ok || role != "admin" {
				return nil, status.Error(codes.PermissionDenied, customErr.ErrAdminPrivilege.Error())
			}
		}

		return handler(ctx, req)
	}
}
