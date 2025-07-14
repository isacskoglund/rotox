// Package probe implements the probe service for the rotox proxy system.
//
// The probe is the component that actually establishes outbound connections
// to target destinations. It receives forwarding requests from the hub via
// gRPC and handles the actual network traffic relaying.
//
// Probes are designed to be lightweight and stateless, making them suitable
// for deployment in serverless environments where they can scale based on
// demand. Each probe can handle multiple concurrent connections.
//
// The probe service:
//   - Receives forward requests from the hub
//   - Establishes connections to target addresses
//   - Relays traffic bidirectionally between client and target
//   - Handles connection errors and cleanup
package probe

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/fault"
)

// Service implements the probe forwarding service.
// It handles requests to establish connections to target addresses
// and relay traffic between the hub and the target.
type Service struct {
	logger *slog.Logger // Logger for probe operations
	dialer netDialer    // Network dialer for establishing outbound connections
}

// NewService creates a new probe service instance with the provided logger and dialer.
// The dialer is used to establish outbound connections to target addresses.
func NewService(
	logger *slog.Logger,
	dialer netDialer,
) *Service {
	return &Service{
		logger: logger,
		dialer: dialer,
	}
}

// Forward handles a forwarding request by establishing a connection to the target
// address and relaying traffic bidirectionally. The accept function is called
// after the target connection is established to get the client connection.
//
// The function establishes the target connection first, then calls accept to
// get the client connection, and finally starts bidirectional traffic relay.
func (svc *Service) Forward(
	ctx context.Context,
	address string,
	accept func() (common.Conn, error),
) error {
	targetTcpConn, err := svc.dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		err = interpretDialError(err)
		return err
	}
	defer targetTcpConn.Close()

	targetConn := &namedConn{
		Conn: targetTcpConn,
		name: "target",
	}

	clientConn, err := accept()
	if err != nil {
		return err
	}

	common.Duplex(
		ctx,
		svc.logger,
		targetConn,
		clientConn,
	)

	return nil
}

// netDialer provides the interface for establishing network connections.
// This abstraction allows for testing with mock dialers and different
// network implementations.
type netDialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

// namedConn wraps a net.Conn with a name for logging and debugging purposes.
type namedConn struct {
	net.Conn
	name string
}

// Name returns the human-readable name of the connection.
func (conn *namedConn) Name() string {
	return conn.name
}

// interpretDialError converts standard network errors into application-specific
// error codes that can be communicated back to the hub and client.
func interpretDialError(err error) error {
	if err == nil {
		return nil
	}
	prefix := "failed to dial"
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fault.Wrap(err, prefix, common.ForwardFailedToResolveHost)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fault.Wrap(err, prefix, common.ForwardHostUnreachable)
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fault.Wrap(err, prefix, common.ForwardHostUnreachable)
	}
	return fault.Wrap(err, prefix, common.ForwardUnknown)
}
