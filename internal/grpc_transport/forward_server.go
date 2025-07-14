package grpc_transport

import (
	"fmt"
	"log/slog"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/fault"
	"github.com/isacskoglund/rotox/internal/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ForwardServer implements the gRPC server side of the ForwardService.
// It receives forwarding requests from hubs and delegates them to the
// underlying forwarder implementation (typically a probe service).
type ForwardServer struct {
	forward_pb.UnimplementedForwardServiceServer
	logger              *slog.Logger     // Logger for server operations
	svc                 common.Forwarder // Underlying forwarding service
	connReadFromBufSize uint             // Buffer size for connection operations
}

// NewForwardServer creates a new gRPC forward server that wraps the provided
// forwarder service. The server handles the gRPC protocol details and
// delegates actual forwarding to the service.
func NewForwardServer(logger *slog.Logger, svc common.Forwarder) *ForwardServer {
	return &ForwardServer{
		logger:              logger,
		svc:                 svc,
		connReadFromBufSize: defaultReadFromBufSize,
	}
}

func (srv *ForwardServer) Forward(stream grpc.BidiStreamingServer[forward_pb.ForwardRequest, forward_pb.ForwardResponse]) error {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceId, ok := md[traceKey]
		if ok && len(traceId) == 1 {
			ctx = tracing.WithTraceId(ctx, traceId[0])
		}
	}
	srv.logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Handling request.",
	)

	msg, err := stream.Recv()
	if err != nil {
		srv.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to receive from stream.",
			slog.Any("error", err),
		)
		return status.Error(codes.Internal, "")
	}

	dialRequest := msg.GetDialRequest()
	if dialRequest == nil {
		srv.logger.LogAttrs(
			ctx,
			slog.LevelWarn,
			"Initial received request was not a dial request. Rejecting.",
		)
		return status.Errorf(
			codes.FailedPrecondition,
			"initial request must be a dial request",
		)
	}
	if dialRequest.Destination == "" {
		return status.Errorf(
			codes.InvalidArgument,
			"destination cannot be empty",
		)
	}

	accept := func() (common.Conn, error) {
		err := stream.Send(&forward_pb.ForwardResponse{
			Response: &forward_pb.ForwardResponse_DialResponse{
				DialResponse: &forward_pb.DialResponse{
					Code: forward_pb.DialResponse_CODE_UNSPECIFIED,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to acknowledge successful dial")
		}
		return newServerConn(stream, "client", srv.connReadFromBufSize), nil
	}

	err = srv.svc.Forward(ctx, dialRequest.Destination, accept)

	switch fault.Code[common.ForwardErrorCode](err) {
	case fault.Ok:
		return nil
	case common.ForwardUnknown:
		return status.Error(codes.Unknown, "")
	case common.ForwardInternal:
		return status.Error(codes.Internal, "")
	}

	var pbCode forward_pb.DialResponse_Code
	switch fault.Code[common.ForwardErrorCode](err) {
	case common.ForwardFailedToResolveHost:
		pbCode = forward_pb.DialResponse_CODE_FAILED_TO_RESOLVE_HOST
	case common.ForwardHostUnreachable:
		pbCode = forward_pb.DialResponse_CODE_HOST_UNREACHABLE
	}
	err = stream.Send(&forward_pb.ForwardResponse{
		Response: &forward_pb.ForwardResponse_DialResponse{
			DialResponse: &forward_pb.DialResponse{
				Code: pbCode,
			},
		},
	})
	if err != nil {
		srv.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to send error to hub.",
			slog.Any("error", err),
		)
		return status.Error(codes.Internal, "")
	}
	return nil
}

// Config
func (srv *ForwardServer) SetReadFromBufSize(size uint) {
	srv.connReadFromBufSize = size
}
