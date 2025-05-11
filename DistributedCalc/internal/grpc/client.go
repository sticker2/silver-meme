package grpc

import (
	"DistributedCalc/pkg/logger"
	"context"
	"time"

	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client CalcServiceClient
	logr   *logger.Logger
}

func NewClient(addr string, logr *logger.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logr.Error("Failed to connect to gRPC server: %v", err)
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: NewCalcServiceClient(conn),
		logr:   logr,
	}, nil
}

func (c *Client) Calculate(ctx context.Context, expr string) (*CalcResponse, error) {
	return c.client.Calculate(ctx, &CalcRequest{Expression: expr})
}

func (c *Client) Close() {
	c.conn.Close()
}

type ClientMock struct {
	CalculateFunc func(ctx context.Context, expr string) (*CalcResponse, error)
}

func (m *ClientMock) Calculate(ctx context.Context, expr string) (*CalcResponse, error) {
	return m.CalculateFunc(ctx, expr)
}

func (m *ClientMock) Close() {}
