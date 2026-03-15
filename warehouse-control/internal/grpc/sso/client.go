package sso

import (
	"context"
	"fmt"
	"warehouse-control/internal/config"

	ssov1 "github.com/sj-shoff/sso_proto/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	authClient ssov1.AuthClient
	conn       *grpc.ClientConn
	cfg        *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
	conn, err := grpc.NewClient(cfg.SSO.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial sso: %w", err)
	}
	return &Client{
		authClient: ssov1.NewAuthClient(conn),
		conn:       conn,
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.SSO.ClientTimeout)
	defer cancel()

	resp, err := c.authClient.Login(ctx, &ssov1.LoginRequest{
		Username: username,
		Password: password,
		AppId:    1,
	})
	if err != nil {
		return "", fmt.Errorf("sso login: %w", err)
	}
	return resp.GetToken(), nil
}
