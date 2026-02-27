package sso

import (
	"context"
	"fmt"
	"time"

	ssov1 "github.com/sj-shoff/sso_proto/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	authClient ssov1.AuthClient
	conn       *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial sso: %w", err)
	}
	return &Client{
		authClient: ssov1.NewAuthClient(conn),
		conn:       conn,
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
