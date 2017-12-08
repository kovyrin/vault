package dbplugin

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes"
)

// ---- gRPC Server domain ----

type gRPCServer struct {
	impl Database
}

func (s *gRPCServer) Type(context.Context, *Empty) (*TypeResponse, error) {
	t, err := s.impl.Type()
	if err != nil {
		return nil, err
	}

	return &TypeResponse{
		Type: t,
	}, nil
}

func (s *gRPCServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	e, err := ptypes.Timestamp(req.Expiration)
	if err != nil {
		return nil, err
	}

	u, p, err := s.impl.CreateUser(ctx, *req.Statements, *req.UsernameConfig, e)

	return &CreateUserResponse{
		Username: u,
		Password: p,
	}, err
}

func (s *gRPCServer) RenewUser(ctx context.Context, req *RenewUserRequest) (*Empty, error) {
	e, err := ptypes.Timestamp(req.Expiration)
	if err != nil {
		return nil, err
	}
	err = s.impl.RenewUser(ctx, *req.Statements, req.Username, e)
	return &Empty{}, err
}

func (s *gRPCServer) RevokeUser(ctx context.Context, req *RevokeUserRequest) (*Empty, error) {
	err := s.impl.RevokeUser(ctx, *req.Statements, req.Username)
	return &Empty{}, err
}

func (s *gRPCServer) Initialize(ctx context.Context, req *InitializeRequest) (*Empty, error) {
	config := map[string]interface{}{}

	err := json.Unmarshal(req.Config, &config)
	if err != nil {
		return nil, err
	}

	err = s.impl.Initialize(ctx, config, req.VerifyConnection)
	return &Empty{}, err
}

func (s *gRPCServer) Close(_ context.Context, _ *Empty) (*Empty, error) {
	s.impl.Close()
	return &Empty{}, nil
}

// ---- gRPC client domain ----

type gRPCClient struct {
	client     DatabaseClient
	clientConn *grpc.ClientConn
}

func (c gRPCClient) Type() (string, error) {
	resp, err := c.client.Type(context.Background(), &Empty{}, grpc.FailFast(true))
	if err != nil {
		return "", err
	}

	return resp.Type, err
}

func (c gRPCClient) CreateUser(ctx context.Context, statements Statements, usernameConfig UsernameConfig, expiration time.Time) (username string, password string, err error) {
	t, err := ptypes.TimestampProto(expiration)
	if err != nil {
		return "", "", err
	}

	resp, err := c.client.CreateUser(ctx, &CreateUserRequest{
		Statements:     &statements,
		UsernameConfig: &usernameConfig,
		Expiration:     t,
	}, grpc.FailFast(true))
	if err != nil {
		return "", "", err
	}

	return resp.Username, resp.Password, err
}

func (c *gRPCClient) RenewUser(ctx context.Context, statements Statements, username string, expiration time.Time) error {
	t, err := ptypes.TimestampProto(expiration)
	if err != nil {
		return err
	}

	_, err = c.client.RenewUser(ctx, &RenewUserRequest{
		Statements: &statements,
		Username:   username,
		Expiration: t,
	}, grpc.FailFast(true))

	return err
}

func (c *gRPCClient) RevokeUser(ctx context.Context, statements Statements, username string) error {
	_, err := c.client.RevokeUser(ctx, &RevokeUserRequest{
		Statements: &statements,
		Username:   username,
	}, grpc.FailFast(true))

	return err
}

func (c *gRPCClient) Initialize(ctx context.Context, config map[string]interface{}, verifyConnection bool) error {
	configRaw, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = c.client.Initialize(ctx, &InitializeRequest{
		Config:           configRaw,
		VerifyConnection: verifyConnection,
	}, grpc.FailFast(true))

	return err
}

func (c *gRPCClient) Close() error {
	_, err := c.client.Close(context.Background(), &Empty{}, grpc.FailFast(true))
	return err
}