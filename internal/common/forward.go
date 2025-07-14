// Package common provides shared interfaces and utilities used across the rotox system.
//
// This package defines the core abstractions for network connections, dialers,
// and forwarding services that are implemented by different transport layers
// (such as gRPC transport) and used by both hub and probe components.
//
// The interfaces are designed to be transport-agnostic, allowing different
// implementations to be plugged in while maintaining a consistent API for
// connection handling and traffic forwarding.
package common

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/isacskoglund/rotox/internal/fault"
)

// ForwardErrorCode represents error types that can occur during forwarding operations.
type ForwardErrorCode string

// Error codes for forwarding operations.
const (
	ForwardUnknown                              = fault.Unknown            // Unknown error occurred
	ForwardInternal            ForwardErrorCode = "INTERNAL"               // Internal system error
	ForwardFailedToResolveHost ForwardErrorCode = "FAILED_TO_RESOLVE_HOST" // DNS resolution failed
	ForwardHostUnreachable     ForwardErrorCode = "HOST_UNREACHABLE"       // Target host is unreachable
)

// Conn represents a network connection with additional metadata.
// Implementing io.WriteTo and io.ReadFrom is beneficial for performance
// when available in the underlying transport.
type Conn interface {
	io.ReadWriteCloser
	Name() string // Returns a human-readable name for logging/debugging
}

// Dialer provides the ability to establish connections to remote addresses.
type Dialer interface {
	// Dial establishes a connection to the specified address.
	// The context can be used to cancel the dial operation.
	Dial(ctx context.Context, address string) (Conn, error)
}

// Forwarder provides the ability to forward connections to target addresses.
// The accept function is called to establish the client connection after
// the target connection has been successfully established.
type Forwarder interface {
	Forward(
		ctx context.Context,
		targetAddress string,
		accept func() (Conn, error),
	) error
}

// Duplex performs bidirectional data transfer between two connections.
// It starts two goroutines to handle data flow in each direction and
// waits for either direction to complete. When one direction ends,
// the context is canceled to signal the other direction to stop.
func Duplex(
	ctx context.Context,
	logger *slog.Logger,
	conn1 Conn,
	conn2 Conn,
) {
	ctx, cancel := context.WithCancel(ctx)

	go simplex(ctx, cancel, logger, conn1, conn2)
	go simplex(ctx, cancel, logger, conn2, conn1)

	<-ctx.Done()
}

// simplex handles unidirectional data transfer from one connection to another.
// It copies data until an error occurs or EOF is reached, then cancels the
// context to signal completion. Normal connection closure errors are not
// logged as errors.
func simplex(
	ctx context.Context,
	cancel func(),
	logger *slog.Logger,
	from Conn,
	to Conn,
) {
	defer cancel()
	n, err := io.Copy(to, from)
	if isAbnormalEror(err) {
		logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Unexpected error when copying data.",
			slog.Int64("writtenBytes", n),
			slog.String("from", from.Name()),
			slog.String("to", to.Name()),
			slog.Any("error", err),
		)
		return
	}
	logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Successfully copied data.",
		slog.Int64("writtenBytes", n),
		slog.String("from", from.Name()),
		slog.String("to", to.Name()),
	)
}

// isAbnormalEror determines if an error represents an abnormal condition
// that should be logged. Normal connection closure errors are not considered abnormal.
func isAbnormalEror(err error) bool {
	if err == nil {
		return false
	}
	if strings.HasSuffix(err.Error(), useOfClosedNetworkConnection) {
		return false
	}

	if errors.Is(err, io.ErrClosedPipe) {
		return false
	}
	return true
}

// useOfClosedNetworkConnection is the error message suffix that indicates
// a connection was closed normally.
const useOfClosedNetworkConnection = "use of closed network connection"
