package client

import (
	"errors"
	"time"

	"github.com/square/quotaservice/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client is a QuotaService client class, adding syntactic sugar over the raw gRPC calls.
type Client struct {
	cc       *grpc.ClientConn
	qsClient quotaservice.QuotaServiceClient
}

// Allow invokes "Allow()" on the "AllowService", taking in a raw AllowRequest message and
// returning the raw AllowResponse message, and optionally any error encountered.
func (c *Client) Allow(request *quotaservice.AllowRequest) (*quotaservice.AllowResponse, error) {
	return c.qsClient.Allow(context.Background(), request)
}

// AllowBlocking adds some syntactic sugar, parsing the response from the QuotaService and blocking,
// if necessary, until the requested quota is available. If this method doesn't return an error
// response, it means quota has been granted and is usable by the time the method returns.
func (c *Client) AllowBlocking(request *quotaservice.AllowRequest) error {
	response, err := c.qsClient.Allow(context.Background(), request)
	if err != nil {
		return err
	}

	if response.Status != quotaservice.AllowResponse_OK {
		// A REJECT response. Return an error.
		return errors.New(quotaservice.AllowResponse_Status_name[int32(response.Status)])
	}

	if response.WaitMillis > 0 {
		time.Sleep(time.Millisecond * time.Duration(response.WaitMillis))
	}

	return nil
}

// Close releases any resources associated with client connections.
func (c *Client) Close() error {
	return c.cc.Close()
}

// New creates a simple new client, connected to a single server. opts can be used to set a
// grpc.Balancer, name resolver, etc. See
// https://godoc.org/google.golang.org/grpc#DialOption for more details.
func New(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{conn, quotaservice.NewQuotaServiceClient(conn)}, nil
}
