package hub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/fault"
	"github.com/isacskoglund/rotox/internal/tracing"
)

// regularDefaultPort is the default port for HTTP connections.
const regularDefaultPort = "80"

// HttpApi implements the HTTP proxy server interface.
// It handles both regular HTTP requests and HTTP CONNECT tunnel requests,
// forwarding them through the hub's probe network.
type HttpApi struct {
	logger *slog.Logger // Logger for HTTP API operations
	core   *Core        // Core hub service for request forwarding
}

// NewHttpApi creates a new HTTP API instance that serves proxy requests.
// The API handles the HTTP protocol specifics and delegates actual
// forwarding to the provided core service.
func NewHttpApi(
	logger *slog.Logger,
	core *Core,
) *HttpApi {
	return &HttpApi{
		logger: logger,
		core:   core,
	}
}

func (api *HttpApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	traceId, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Errorf("failed to randomize uuid: %w", err))
	}
	ctx := tracing.WithTraceId(req.Context(), traceId.String())
	req = req.WithContext(ctx)
	conn, err := hijack(w, "client")
	if err != nil {
		api.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Error when hijacking connection",
			slog.Any("error", err),
		)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	if req.Method == "CONNECT" {
		api.handleConnect(conn, req)
	} else {
		api.handlePlain(conn, req)
	}

}

func (api *HttpApi) handleConnect(conn common.Conn, req *http.Request) {
	ctx := req.Context()
	api.logger.LogAttrs(
		ctx,
		slog.LevelInfo,
		"Handling CONNECT request",
	)

	accept := func() (common.Conn, error) {
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			return nil, fmt.Errorf("failed to write 200 OK to client: %w", err)
		}
		return conn, nil
	}

	err := api.core.forward(
		ctx,
		req.URL.Host,
		accept,
	)
	api.handleForwardError(ctx, conn, err)
}

func (api *HttpApi) handlePlain(conn common.Conn, req *http.Request) {
	ctx := req.Context()
	api.logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Handling regular request",
	)

	req.Body = http.NoBody
	var headBuf bytes.Buffer
	if err := req.WriteProxy(&headBuf); err != nil {
		api.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to write req (without body) to buffer",
			slog.Any("error", err),
		)
		writeHttpError(conn, http.StatusInternalServerError)
		return
	}
	reader := io.MultiReader(&headBuf, conn)

	wrappedConn := customConn{
		Reader: reader,
		Writer: conn,
		Closer: conn,
		namer:  conn,
	}

	accept := func() (common.Conn, error) {
		return wrappedConn, nil
	}

	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, regularDefaultPort)
	}
	err := api.core.forward(ctx, host, accept)
	api.handleForwardError(ctx, conn, err)
}

func (api *HttpApi) handleForwardError(ctx context.Context, conn common.Conn, err error) {
	if err == nil {
		return
	}
	switch fault.Code[common.ForwardErrorCode](err) {
	case common.ForwardUnknown, common.ForwardInternal:
		api.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Unknown error when forwarding connection",
			slog.Any("error", err),
		)
		writeHttpError(conn, http.StatusInternalServerError)
	case common.ForwardFailedToResolveHost:
		api.logger.LogAttrs(
			ctx,
			slog.LevelInfo,
			"Failed to resolve target host when forwarding connection.",
			slog.Any("error", err),
		)
		writeHttpError(conn, 502)
	case common.ForwardHostUnreachable:
		api.logger.LogAttrs(
			ctx,
			slog.LevelInfo,
			"Failed to reach target when forwarding connection.",
			slog.Any("error", err),
		)
		writeHttpError(conn, 504)
	}

	if err != nil {
		// TODO: If this is a dial error (unknown host), it should not be 500.
		api.logger.LogAttrs(
			ctx,
			slog.LevelError,
			"Failed to forward regular connection",
			slog.Any("error", err),
		)
	}

}

type namer interface {
	Name() string
}

type customConn struct {
	namer
	io.Reader
	io.Writer
	io.Closer
}

type customNamer struct {
	name string
}

func (n *customNamer) Name() string {
	return n.name
}

func hijack(w http.ResponseWriter, name string) (common.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("failed to type assert w as http.Hijacker")
	}
	tcpConn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("failed to hijack connection: %w", err)
	}
	return &customConn{
		Reader: io.MultiReader(buf, tcpConn),
		Writer: tcpConn,
		Closer: tcpConn,
		namer:  &customNamer{name: name},
	}, nil
}

func writeHttpError(w io.Writer, statusCode int) error {
	statusText := http.StatusText(statusCode)
	resp := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
		statusCode,
		statusText,
	)
	_, err := io.WriteString(w, resp)
	return err
}
