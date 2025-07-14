// Package hub implements the central coordinator of the rotox proxy system.
//
// The hub receives incoming proxy requests from clients and distributes them
// across a pool of probes using round-robin load balancing. It manages the
// lifecycle of connections and provides telemetry about traffic flow.
//
// The hub is responsible for:
//   - Accepting HTTP proxy requests (both plaintext and CONNECT)
//   - Load balancing across available probes
//   - Managing connection lifecycle
//   - Publishing telemetry events
//   - Handling probe failures and retries
package hub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/telemetry"
)

// Core represents the central hub service that coordinates proxy requests
// across multiple probes. It implements round-robin load balancing and
// provides telemetry integration.
type Core struct {
	logger *slog.Logger             // Logger for hub operations
	tel    *multiTelemetryPublisher // Telemetry publisher for events
	probes []common.Dialer          // Pool of available probes
	next   chan int                 // Channel for round-robin probe selection
}

// NewCore creates a new hub core instance with the provided logger and probes.
// At least one probe must be provided or the function will panic.
// The core uses round-robin load balancing to distribute requests across probes.
func NewCore(
	logger *slog.Logger,
	probes []common.Dialer,
) *Core {
	if len(probes) == 0 {
		panic("Probes must not be empty")
	}

	var next = make(chan int, 1)
	next <- 0 // Initialize with starting value

	return &Core{
		logger: logger,
		probes: probes,
		tel:    newMultiTelemetryPublisher(),
		next:   next,
	}
}

// RegisterTelemetryDispatcher adds a telemetry publisher to receive hub events.
// Multiple publishers can be registered to send telemetry to different destinations.
func (core *Core) RegisterTelemetryDispatcher(dis telemetryPublisher) {
	core.tel.register(dis)
}

// forward handles a single proxy request by selecting a probe and establishing
// the necessary connections. It uses round-robin load balancing to select the next
// available probe and publishes telemetry events about the connection lifecycle.
func (core *Core) forward(
	ctx context.Context,
	targetAddress string,
	accept func() (common.Conn, error),
) error {
	// Get the next probe to be used (round robin)
	probeIdx := <-core.next
	if probeIdx >= len(core.probes)-1 {
		core.next <- 0
	} else {
		core.next <- probeIdx + 1
	}
	probe := core.probes[probeIdx]

	core.logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Forwarding connection.",
		slog.Int("probeIdx", probeIdx),
	)

	// Dial to the target
	targetConn, err := probe.Dial(ctx, targetAddress)
	if err != nil {
		return err
	}
	defer targetConn.Close()

	// Accept the client connection
	// (only once connection to target has been established)
	clientConn, err := accept()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	connectionId, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to create random connection id: %w", err)
	}

	// Relay the traffic (bidirectional)
	emitTransferEvent := func(start time.Time, stop time.Time, n uint64) {
		core.tel.TransferPublisher().Publish(
			telemetry.TransferEvent{
				ConnectionId: connectionId.String(),
				StartedAt:    start,
				FinishedAt:   stop,
				BytesCount:   n,
			},
		)
	}
	openedAt := time.Now()
	core.tel.ConnectionPublisher().Publish(
		telemetry.ConnectionEvent{
			ConnectionId:  connectionId.String(),
			ClientAddress: "not set",
			TargetAddress: targetAddress,
			OpenedAt:      openedAt,
			ClosedAt:      time.Unix(0, 0),
		},
	)
	defer core.tel.ConnectionPublisher().Publish(
		telemetry.ConnectionEvent{
			ConnectionId:  connectionId.String(),
			ClientAddress: "not set",
			TargetAddress: targetAddress,
			OpenedAt:      openedAt,
			ClosedAt:      time.Now(),
		},
	)
	common.Duplex(
		ctx,
		core.logger,
		telemetry.NewTelemetryConn(targetConn, emitTransferEvent),
		telemetry.NewTelemetryConn(clientConn, emitTransferEvent),
	)

	core.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Connection closed",
	)
	return nil
}
