// Package grpc_transport provides gRPC-based transport implementations for rotox.
//
// This package implements the common.Dialer and common.Forwarder interfaces
// using gRPC as the transport mechanism. It handles:
//   - Client-side connection establishment to probes
//   - Server-side handling of forwarding requests
//   - Bidirectional streaming for data transfer
//   - Authentication and authorization
//   - Connection lifecycle management
//   - Tracing integration
//
// The transport layer abstracts the complexity of gRPC streaming and provides
// a simple connection-like interface that can be used with standard Go
// networking patterns.
package grpc_transport

import (
	"context"
	"fmt"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/fault"
	"github.com/isacskoglund/rotox/internal/tracing"
	"google.golang.org/grpc/metadata"
)

// traceKey is the metadata key used for propagating trace IDs in gRPC calls.
const traceKey = "trace_id"

// forwardClient implements common.Dialer using gRPC to communicate with probes.
// It establishes connections through the ForwardService gRPC service.
type forwardClient struct {
	client              forward_pb.ForwardServiceClient // gRPC client for probe communication
	connReadFromBufSize uint                            // Buffer size for connection reads
}

// NewForwardClient creates a new gRPC-based dialer for connecting to probes.
// The client should be configured with appropriate authentication and transport settings.
func NewForwardClient(
	client forward_pb.ForwardServiceClient,
) common.Dialer {
	return &forwardClient{
		client:              client,
		connReadFromBufSize: defaultReadFromBufSize,
	}
}

// Dial establishes a connection to the specified address through a probe.
// It creates a gRPC stream, sends a dial request, waits for confirmation,
// and returns a connection that can be used for data transfer.
func (dialer *forwardClient) Dial(ctx context.Context, address string) (common.Conn, error) {
	stream, err := dialer.client.Forward(
		metadata.AppendToOutgoingContext(
			ctx,
			traceKey,
			tracing.GetTraceId(ctx),
		),
	)
	if err != nil {
		return nil, err
	}
	err = stream.Send(&forward_pb.ForwardRequest{
		Request: &forward_pb.ForwardRequest_DialRequest{
			DialRequest: &forward_pb.DialRequest{
				Destination: address,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	msg, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	dialResponse := msg.GetDialResponse()
	if dialResponse == nil {
		return nil, fmt.Errorf("dial response was nil")
	}
	switch dialResponse.Code {
	case forward_pb.DialResponse_CODE_UNSPECIFIED:
		return newClientConn(stream, "target", dialer.connReadFromBufSize), nil
	case forward_pb.DialResponse_CODE_FAILED_TO_RESOLVE_HOST:
		return nil, fault.New("failed to resolve host", common.ForwardFailedToResolveHost)
	case forward_pb.DialResponse_CODE_HOST_UNREACHABLE:
		return nil, fault.New("host unreachable", common.ForwardHostUnreachable)
	}
	panic("unreachable")
}

func (client *forwardClient) SetReadFromBufSize(size uint) {
	client.connReadFromBufSize = size
}
