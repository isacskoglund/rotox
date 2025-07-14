// Package main implements the rotox hub server.
//
// The hub is the central coordinator in the rotox distributed proxy.
// It receives incoming proxy requests from clients and forwards them to one
// of the available probes, distributing the traffic over a pool public IP addresses.
//
// The hub supports two proxy protocols:
//   - HTTP plaintext requests: Direct forwarding of HTTP requests
//   - HTTP CONNECT requests: Tunneling for HTTPS and other protocols
//
// Configuration is loaded from a YAML file specified by the CONFIG_FILE
// environment variable. The hub can manage multiple probe groups with
// different authentication secrets and TLS requirements.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	telemetry_pb "github.com/isacskoglund/rotox/gen/go/telemetry/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/config"
	"github.com/isacskoglund/rotox/internal/grpc_transport"
	"github.com/isacskoglund/rotox/internal/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

// ProbeConfig represents the configuration for a group of probes.
// Each probe group can have multiple hosts with shared settings.
type ProbeConfig struct {
	SecretEnv  *string  `yaml:"secret_env" validate:"omitempty,envexists"` // Environment variable containing the probe secret
	RequireTls *bool    `yaml:"require_tls" validate:"required"`           // Whether TLS is required for probe connections
	Hosts      []string `yaml:"hosts" validate:"required,min=1,dive"`      // The address of each probe in this group
}

// Config represents the complete hub configuration loaded from YAML.
type Config struct {
	LogLevel  string `yaml:"log_level" validate:"required,oneof=debug info warn error"` // Logging verbosity level
	LogFormat string `yaml:"log_format" validate:"required,oneof=json text"`            // Log output format

	Proxies struct {
		Http *struct {
			SecretEnv *string `yaml:"secret_env" validate:"omitempty,envexists"` // Environment variable for HTTP proxy secret
			Port      int     `yaml:"port" validate:"required,min=1,max=65535"`  // Port for HTTP proxy listener
		} `yaml:"http"`
	} `yaml:"proxies"`

	Telemetry *struct {
		SecretEnv *string `yaml:"secret_env" validate:"omitempty,envexists"` // Environment variable for telemetry secret
		Port      int     `yaml:"port" validate:"required,min=1,max=65535"`  // Port for telemetry server
	} `yaml:"telemetry"`

	Probes []ProbeConfig `yaml:"probes" validate:"required,min=1,dive"` // List of probe configurations
}

// main initializes and starts the rotox hub server.
// It loads configuration, sets up logging, initializes probes and services,
// and starts both the HTTP proxy server and optionally the telemetry server.
func main() {
	ctx := context.Background()
	cfg, err := LoadConfig("config/hub.yaml")
	if err != nil {
		panic(fmt.Errorf("error loading config: %w", err))
	}

	logger, err := config.LoggerFromConfig(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}
	probes := setupProbes(cfg.Probes)
	core := hub.NewCore(logger, probes)
	telemetrySrv := grpc_transport.NewTelemetryServer(logger)
	core.RegisterTelemetryDispatcher(telemetrySrv)
	httpApi := hub.NewHttpApi(logger, core)

	logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Starting hub server",
		slog.Any("config", cfg),
	)

	if cfg.Telemetry != nil {
		var opts []grpc.ServerOption
		if cfg.Telemetry.SecretEnv != nil {
			opts = append(
				opts,
				grpc.StreamInterceptor(
					grpc_transport.NewServerInterceptor(
						logger,
						os.Getenv(*cfg.Telemetry.SecretEnv),
					),
				),
			)
		}
		grpcSrv := grpc.NewServer(opts...)
		telemetry_pb.RegisterTelemetryServiceServer(grpcSrv, telemetrySrv)
		telemetrySrv.StartBroadcasting(ctx)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Telemetry.Port))
		if err != nil {
			log.Fatal("Failed to listen to grpc port", err)
		}
		go grpcSrv.Serve(lis)
	}

	if cfg.Proxies.Http == nil {
		log.Fatalf("no proxies are enabled")
	}

	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Proxies.Http.Port), httpApi)
	if err != nil {
		log.Fatalf("error when listening: %v", err)
	}
}

// setupProbes creates dialer instances for all configured probes.
// It iterates through all probe configurations and creates a separate
// dialer for each host in each probe group.
func setupProbes(cfg []ProbeConfig) []common.Dialer {
	probes := []common.Dialer{}
	for _, probe := range cfg {
		for _, host := range probe.Hosts {
			probes = append(
				probes,
				grpc_transport.NewForwardClient(
					setupProbeClient(
						host,
						os.Getenv(*probe.SecretEnv),
						*probe.RequireTls,
					),
				),
			)
		}
	}
	return probes
}

// setupProbeClient creates a gRPC client for connecting to a probe.
// It handles hostname normalization, TLS configuration, and authentication.
func setupProbeClient(hostname string, secret string, requireTls bool) forward_pb.ForwardServiceClient {
	hostname = strings.Replace(hostname, "http://", "dns:///", 1)
	hostname = strings.Replace(hostname, "https://", "dns:///", 1)

	var opts []grpc.DialOption
	if requireTls {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if secret != "" {
		opts = append(opts, grpc.WithStreamInterceptor(grpc_transport.NewClientInterceptor(secret)))
	}
	client, err := grpc.NewClient(hostname, opts...)
	if err != nil {
		log.Fatalf("error connecting to relay: %v", err)
	}
	return forward_pb.NewForwardServiceClient(client)
}

// LogValue implements slog.LogValuer for structured logging of configuration.
// It returns a slog.Value containing the essential configuration parameters
// while avoiding sensitive information like secrets.
func (cfg Config) LogValue() slog.Value {
	probeHostsHead := []string{}
	for _, probes := range cfg.Probes {
		probeHostsHead = append(probeHostsHead, probes.Hosts...)
	}

	return slog.GroupValue(
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
		slog.Any("probe_hosts", probeHostsHead),
	)
}

// LoadConfig loads and validates the hub configuration from a YAML file.
// It uses the CONFIG_FILE environment variable if set, otherwise falls back
// to the provided default filename. The configuration is validated against
// the struct tags and custom validators.
func LoadConfig(defaultConfigFilename string) (*Config, error) {
	cfg := &Config{
		LogLevel:  "debug",
		LogFormat: "json",
	}

	configFilename := defaultConfigFilename
	if f := os.Getenv("CONFIG_FILE"); f != "" {
		configFilename = f
	}

	data, err := os.ReadFile(configFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFilename, err)
	}

	validate := validator.New()
	validate.RegisterValidation("envexists", EnvVarExists)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml config file: %w", err)
	}
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	return cfg, nil
}

// EnvVarExists is a custom validator function that checks if an environment
// variable exists and is not empty. It supports both string and *string field types.
func EnvVarExists(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()
	if kind != reflect.String && !(kind == reflect.Ptr && field.Type().Elem().Kind() == reflect.String) {
		panic(fmt.Sprintf("envexists validation: unsupported field type %s, must be string or *string", field.Type()))
	}

	var envVarName string
	if kind == reflect.Ptr {
		if field.IsNil() {
			fmt.Println("envexists validation: field is nil pointer (empty), but env var expected")
			return false
		}
		envVarName = field.Elem().String()
	} else {
		envVarName = field.String()
	}

	if envVarName == "" {
		fmt.Println("envexists validation: empty env var name provided but required")
		return false
	}

	val, exists := os.LookupEnv(envVarName)
	if !exists || val == "" {
		fmt.Printf("envexists validation: environment variable %q not set or empty\n", envVarName)
		return false
	}

	return true
}
