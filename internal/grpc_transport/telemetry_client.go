package grpc_transport

import (
	"context"
	"time"

	telemetry_pb "github.com/isacskoglund/rotox/gen/go/telemetry/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/telemetry"
	"google.golang.org/grpc"
)

type TelemetryClient struct {
	transferEvents   *grpcTransferSubscriber
	connectionEvents *grpcConnectionSubscriber
}

func NewTelemetryClient(
	client telemetry_pb.TelemetryServiceClient,
) *TelemetryClient {
	return &TelemetryClient{
		transferEvents: &grpcTransferSubscriber{
			client: client,
		},
		connectionEvents: &grpcConnectionSubscriber{
			client: client,
		},
	}
}

func (client *TelemetryClient) TransferSubscriber() common.Subscriber[telemetry.TransferEvent] {
	return client.transferEvents
}
func (client *TelemetryClient) ConnectionSubscriber() common.Subscriber[telemetry.ConnectionEvent] {
	return client.connectionEvents
}

type grpcTransferSubscriber struct {
	client telemetry_pb.TelemetryServiceClient
}

func (s *grpcTransferSubscriber) Subscribe(ctx context.Context) (common.Subscription[telemetry.TransferEvent], error) {
	stream, err := s.client.TransferSubscribe(ctx, &telemetry_pb.TransferSubscribeRequest{})
	if err != nil {
		return nil, err
	}

	convert := func(resp *telemetry_pb.TransferSubscribeResponse) ([]telemetry.TransferEvent, error) {
		converted := make([]telemetry.TransferEvent, len(resp.Events))
		for i, event := range resp.Events {
			converted[i] = telemetry.TransferEvent{
				ConnectionId: event.ConnectionId,
				StartedAt:    time.Unix(0, int64(event.StartedAt)),
				FinishedAt:   time.Unix(0, int64(event.FinishedAt)),
				BytesCount:   event.BytesCount,
			}
		}
		return converted, nil
	}

	return &grpcServerStreamSubscription[telemetry_pb.TransferSubscribeResponse, telemetry.TransferEvent]{
		stream:  stream,
		cache:   make([]telemetry.TransferEvent, 0),
		convert: convert,
	}, nil
}

type grpcConnectionSubscriber struct {
	client telemetry_pb.TelemetryServiceClient
}

func (s *grpcConnectionSubscriber) Subscribe(ctx context.Context) (common.Subscription[telemetry.ConnectionEvent], error) {
	stream, err := s.client.ConnectionSubscribe(ctx, &telemetry_pb.ConnectionSubscribeRequest{})
	if err != nil {
		return nil, err
	}

	convert := func(resp *telemetry_pb.ConnectionSubscribeResponse) ([]telemetry.ConnectionEvent, error) {
		converted := make([]telemetry.ConnectionEvent, len(resp.Events))
		for i, event := range resp.Events {
			converted[i] = telemetry.ConnectionEvent{
				ConnectionId:  event.ConnectionId,
				ClientAddress: event.ClientAddress,
				TargetAddress: event.TargetAddress,
				OpenedAt:      time.Unix(0, int64(event.OpenedAt)),
				ClosedAt:      time.Unix(0, int64(event.ClosedAt)),
			}
		}
		return converted, nil
	}

	return &grpcServerStreamSubscription[telemetry_pb.ConnectionSubscribeResponse, telemetry.ConnectionEvent]{
		stream:  stream,
		cache:   make([]telemetry.ConnectionEvent, 0),
		convert: convert,
	}, nil
}

// Generic subscription interface for gRPC server streaming
type grpcServerStreamSubscription[M any, T any] struct {
	stream  grpc.ServerStreamingClient[M]
	convert func(*M) ([]T, error)
	cache   []T
}

func (sub *grpcServerStreamSubscription[M, T]) Receive() (T, error) {
	err := sub.receive()
	if err != nil {
		return *new(T), err
	}
	event := sub.cache[len(sub.cache)-1]
	sub.cache = sub.cache[:len(sub.cache)-1]
	return event, nil
}

func (sub *grpcServerStreamSubscription[M, T]) Close() {
	sub.stream.CloseSend()
}

func (sub *grpcServerStreamSubscription[M, T]) receive() error {
	if len(sub.cache) > 0 {
		return nil
	}
	for {
		resp, err := sub.stream.Recv()
		if err != nil {
			return err
		}
		converted, err := sub.convert(resp)
		if err != nil {
			return err
		}
		sub.cache = append(sub.cache, converted...)
		if len(sub.cache) > 0 {
			return nil
		}
	}
}
