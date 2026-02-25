package authgrpc

import "context"

type auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		username string,
		password string,
	) (userID int64, err error)
}
