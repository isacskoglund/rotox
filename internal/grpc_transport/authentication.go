package grpc_transport

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// NewServerInterceptor creates a gRPC server interceptor that validates
// authentication using a shared secret. The secret must be provided in
// the "authorization" metadata field with "Bearer " prefix.
//
// TODO: Improve logging and extract peer information with peer.FromContext
// TODO: or from "x-forwarded-for" header depending on deployment.
func NewServerInterceptor(
	logger *slog.Logger,
	secret string,
) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// TODO:
		// Better logging.
		// Extract peer information with peer.FromContext or from the "x-forwarded-for", depending on deployment.
		ctx := ss.Context()
		logger.LogAttrs(
			ctx,
			slog.LevelDebug,
			"authenticating incoming request",
		)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}
		header, ok := md["authorization"]
		if !ok {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		if len(header) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization token")
		}

		bearer, token, found := strings.Cut(header[0], " ")
		if !found || strings.ToLower(bearer) != "bearer" {
			return status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		if token != secret {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(srv, ss)
	}
}

// NewClientInterceptor creates a gRPC client interceptor that automatically
// adds authentication metadata to outgoing requests. The token is added
// as a "Bearer" token in the "authorization" metadata field.
func NewClientInterceptor(
	token string,
) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Attach authorization metadata to the context
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+token)

		// Proceed with the streaming call using the modified context
		return streamer(ctx, desc, cc, method, opts...)
	}
}
