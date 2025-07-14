// Package telemetry provides event types and interfaces for monitoring rotox traffic.
//
// This package defines the events that are published throughout the system
// to provide visibility into connection lifecycle, data transfer, and system
// performance. Events can be consumed by various telemetry backends for
// monitoring, alerting, and analytics.
package telemetry

import "time"

// ConnectionEvent represents the lifecycle of a single proxy connection.
// It tracks when connections are opened and closed, along with the
// client and target addresses involved.
type ConnectionEvent struct {
	ConnectionId  string    // Unique identifier for the connection
	ClientAddress string    // Address of the connecting client
	TargetAddress string    // Address of the target destination
	OpenedAt      time.Time // When the connection was established
	// ClosedAt indicates when the connection was closed.
	// A zero value indicates the connection is still open.
	ClosedAt time.Time
}

// TransferEvent represents a data transfer operation within a connection.
// It provides metrics about the amount of data transferred and timing.
type TransferEvent struct {
	ConnectionId string    // Connection this transfer belongs to
	StartedAt    time.Time // When the transfer began
	FinishedAt   time.Time // When the transfer completed
	BytesCount   uint64    // Number of bytes transferred
}
