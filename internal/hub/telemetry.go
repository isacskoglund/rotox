package hub

import (
	"errors"

	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/telemetry"
)

type telemetryPublisher interface {
	TransferPublisher() common.Publisher[telemetry.TransferEvent]
	ConnectionPublisher() common.Publisher[telemetry.ConnectionEvent]
}

type multiPublisher[T any] struct {
	publishers []common.Publisher[T]
}

func (m *multiPublisher[T]) Publish(event T) error {
	errs := make([]error, len(m.publishers))
	for i, next := range m.publishers {
		errs[i] = next.Publish(event)
	}
	return errors.Join(errs...)
}

func (m *multiPublisher[T]) register(pub common.Publisher[T]) {
	m.publishers = append(m.publishers, pub)
}

type multiTelemetryPublisher struct {
	transferEvents   *multiPublisher[telemetry.TransferEvent]
	connectionEvents *multiPublisher[telemetry.ConnectionEvent]
}

func newMultiTelemetryPublisher() *multiTelemetryPublisher {
	return &multiTelemetryPublisher{
		transferEvents:   &multiPublisher[telemetry.TransferEvent]{},
		connectionEvents: &multiPublisher[telemetry.ConnectionEvent]{},
	}
}

func (mtp *multiTelemetryPublisher) TransferPublisher() common.Publisher[telemetry.TransferEvent] {
	return mtp.transferEvents
}

func (mtp *multiTelemetryPublisher) ConnectionPublisher() common.Publisher[telemetry.ConnectionEvent] {
	return mtp.connectionEvents
}

func (mtp *multiTelemetryPublisher) register(pub telemetryPublisher) {
	mtp.transferEvents.register(pub.TransferPublisher())
	mtp.connectionEvents.register(pub.ConnectionPublisher())
}
