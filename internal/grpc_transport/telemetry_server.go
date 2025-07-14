package grpc_transport

import (
	"context"
	"log/slog"

	telemetry_pb "github.com/isacskoglund/rotox/gen/go/telemetry/v1"
	"github.com/isacskoglund/rotox/internal/broadcast"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/telemetry"
	"google.golang.org/grpc"
)

type TelemetryServer struct {
	telemetry_pb.UnimplementedTelemetryServiceServer
	logger           *slog.Logger
	transferEvents   *broadcast.Broadcaster[telemetry.TransferEvent]
	connectionEvents *broadcast.Broadcaster[telemetry.ConnectionEvent]
}

func NewTelemetryServer(
	logger *slog.Logger,
) *TelemetryServer {
	return &TelemetryServer{
		logger:           logger,
		transferEvents:   broadcast.NewBroadcaster[telemetry.TransferEvent](),
		connectionEvents: broadcast.NewBroadcaster[telemetry.ConnectionEvent](),
	}
}

func (srv *TelemetryServer) StartBroadcasting(ctx context.Context) error {
	err := srv.transferEvents.Start(ctx)
	if err != nil {
		return err
	}
	err = srv.connectionEvents.Start(ctx)
	return err
}

func (srv *TelemetryServer) TransferPublisher() common.Publisher[telemetry.TransferEvent] {
	return srv.transferEvents
}

func (srv *TelemetryServer) ConnectionPublisher() common.Publisher[telemetry.ConnectionEvent] {
	return srv.connectionEvents
}

func (srv *TelemetryServer) TransferSubscribe(req *telemetry_pb.TransferSubscribeRequest, stream grpc.ServerStreamingServer[telemetry_pb.TransferSubscribeResponse]) error {
	ctx := stream.Context()
	srv.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Handling transfer subscribe request.",
	)

	sub, err := srv.transferEvents.Subscribe(ctx)
	if err != nil {
		srv.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to subscribe to transfer events.",
			slog.String("error", err.Error()),
		)
		return err
	}
	defer sub.Close()
	for {
		event, err := sub.Receive()
		if err != nil {
			srv.logger.LogAttrs(
				ctx,
				slog.LevelError,
				"Failed to receive transfer event.",
				slog.String("error", err.Error()),
			)
			return err
		}

		// TODO: Add batching to not send as many messages
		err = stream.Send(
			&telemetry_pb.TransferSubscribeResponse{
				Events: []*telemetry_pb.TransferEvent{
					{
						ConnectionId: event.ConnectionId,
						StartedAt:    uint64(event.StartedAt.UnixNano()),
						FinishedAt:   uint64(event.FinishedAt.UnixNano()),
						BytesCount:   event.BytesCount,
					},
				},
			},
		)
		if err != nil {
			srv.logger.LogAttrs(
				ctx,
				slog.LevelError,
				"Failed to send transfer event.",
				slog.String("error", err.Error()),
			)
			return err
		}

	}
}

func (srv *TelemetryServer) ConnectionSubscribe(req *telemetry_pb.ConnectionSubscribeRequest, stream grpc.ServerStreamingServer[telemetry_pb.ConnectionSubscribeResponse]) error {
	ctx := stream.Context()
	srv.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Handling connection subscribe request.",
	)

	sub, err := srv.connectionEvents.Subscribe(ctx)
	if err != nil {
		srv.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to subscribe to connection events.",
			slog.String("error", err.Error()),
		)
		return err
	}
	defer sub.Close()
	for {
		event, err := sub.Receive()
		if err != nil {
			srv.logger.LogAttrs(
				ctx,
				slog.LevelError,
				"Failed to receive connection event.",
				slog.String("error", err.Error()),
			)
			return err
		}

		// TODO: Add batching to not send as many messages
		err = stream.Send(
			&telemetry_pb.ConnectionSubscribeResponse{
				Events: []*telemetry_pb.ConnectionEvent{
					{
						ConnectionId:  event.ConnectionId,
						OpenedAt:      uint64(event.OpenedAt.UnixNano()),
						ClosedAt:      uint64(event.ClosedAt.UnixNano()),
						ClientAddress: event.ClientAddress,
						TargetAddress: event.TargetAddress,
					},
				},
			},
		)
		if err != nil {
			srv.logger.LogAttrs(
				ctx,
				slog.LevelError,
				"Failed to send connection event.",
				slog.String("error", err.Error()),
			)
			return err
		}

	}

}
