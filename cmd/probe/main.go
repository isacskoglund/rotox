// Package main implements the rotox probe server.
//
// The probe is a component in the rotox distributed proxy that acts
// as the actual egress point for network traffic. A probe receives forward
// requests from the hub via gRPC and establish outbound connections to target
// destinations, relaying traffic bidirectionally.
//
// Probes are designed to be lightweight and stateless, making them suitable
// for deployment in serverless environments like Google Cloud Run where they
// can scale to zero when not in use.
//
// Configuration is provided via environment variables for simplicity in
// containerized deployments.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/config"
	"github.com/isacskoglund/rotox/internal/grpc_transport"
	"github.com/isacskoglund/rotox/internal/probe"
	"github.com/sethvargo/go-envconfig"
	"google.golang.org/grpc"
)

// dialTimeout is the maximum time to wait when establishing outbound connections.
const dialTimeout = 2 * time.Second

// Config represents the probe configuration loaded from environment variables.
type Config struct {
	LogLevel  string  `env:"LOG_LEVEL, default=debug"` // Logging verbosity level
	LogFormat string  `env:"LOG_FORMAT, default=json"` // Log output format (json or text)
	Port      uint16  `env:"PORT, default=8000"`       // Port for the gRPC server to listen on
	Secret    *string `env:"SECRET, required"`         // Authentication secret for hub connections
}

// main initializes and starts the rotox probe server.
// It loads configuration from environment variables, sets up logging,
// creates the probe service, and starts the gRPC server.
func main() {
	ctx := context.Background()
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("error loading config from environment: %v", err)
	}

	logger, err := config.LoggerFromConfig(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	svc := probe.NewService(
		logger,
		&net.Dialer{
			Timeout: dialTimeout,
		},
	)
	srv := grpc_transport.NewForwardServer(logger, svc)

	var opts []grpc.ServerOption
	if cfg.Secret != nil {
		opts = append(opts,
			grpc.StreamInterceptor(
				grpc_transport.NewServerInterceptor(
					logger,
					*cfg.Secret,
				),
			),
		)
	}
	s := grpc.NewServer(opts...)
	forward_pb.RegisterForwardServiceServer(s, srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to listen",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Starting probe server",
		slog.Any("config", cfg),
	)
	if err := s.Serve(lis); err != nil {
		logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to serve",
		)
		os.Exit(1)
	}
}

// LogValue implements slog.LogValuer for structured logging of configuration.
// It returns essential configuration parameters while avoiding sensitive
// information like the actual secret value.
func (cfg Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
		slog.Int("port", int(cfg.Port)),
		slog.Bool("authentication_enabled", cfg.Secret != nil),
	)
}
