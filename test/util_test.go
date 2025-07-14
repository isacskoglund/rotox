package integration_test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/grpc_transport"
	"github.com/isacskoglund/rotox/internal/hub"
	"github.com/isacskoglund/rotox/internal/probe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func serveProbe(lis net.Listener, logger *slog.Logger, dialer *mockDialer) {
	probe := grpc_transport.NewForwardServer(
		logger,
		probe.NewService(logger, dialer),
	)
	probe.SetReadFromBufSize(10)
	s := grpc.NewServer()
	forward_pb.RegisterForwardServiceServer(s, probe)
	go func() {
		err := s.Serve(lis)
		if err.Error() == "closed" {
			return
		}
		if err != nil {
			log.Fatalf("error when serving probe server: %#v, %T", err, err)
		}
	}()
}

func serveHub(httpApiHub *bufconn.Listener, logger *slog.Logger, probes []*bufconn.Listener) {
	dialers := make([]common.Dialer, len(probes))
	for i := range dialers {
		client := grpc_transport.NewForwardClient(
			forward_pb.NewForwardServiceClient(
				newGrpcClient(probes[i]),
			),
		)
		setter := client.(interface{ SetReadFromBufSize(size uint) })
		setter.SetReadFromBufSize(10)
		dialers[i] = client
	}
	core := hub.NewCore(logger, dialers)
	httpApi := hub.NewHttpApi(logger, core)
	go func() {
		err := http.Serve(httpApiHub, httpApi)
		if err != nil && err.Error() != "closed" {
			log.Fatalf("unexpected error when serving http api: %v", err)
		}
	}()
}

func newGrpcClient(lis *bufconn.Listener) *grpc.ClientConn {
	client, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(fmt.Errorf("failed to create client from listener: %w", err))
	}
	return client
}
