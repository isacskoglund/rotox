package grpc_transport

import (
	"fmt"
	"io"
	"net"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// defaultReadFromBufSize is the default buffer size for ReadFrom operations.
const defaultReadFromBufSize = 32 * 1024

// grpcConn implements common.Conn using gRPC bidirectional streams.
// It provides a connection-like interface over gRPC streaming, with
// buffering for efficient data transfer operations.
type grpcConn struct {
	stream          bidiStream // Underlying gRPC stream
	closed          bool       // Connection state
	buf             []byte     // Internal buffer for reads
	offset          int        // Current buffer offset
	readFromBufSize uint       // Buffer size for ReadFrom operations
	name            string     // Connection name for logging
}

// newClientConn creates a new client-side gRPC connection wrapper.
// It wraps a gRPC bidirectional streaming client to provide a
// standard connection interface.
func newClientConn(
	stream grpc.BidiStreamingClient[forward_pb.ForwardRequest, forward_pb.ForwardResponse],
	name string,
	readFromBufSize uint,
) common.Conn {
	return &grpcConn{
		stream:          &bidiClientStream{stream: stream},
		closed:          false,
		buf:             []byte{},
		offset:          0,
		readFromBufSize: readFromBufSize,
		name:            name,
	}
}

// newServerConn creates a new server-side gRPC connection wrapper.
// It wraps a gRPC bidirectional streaming server to provide a
// standard connection interface.
func newServerConn(
	stream grpc.BidiStreamingServer[forward_pb.ForwardRequest, forward_pb.ForwardResponse],
	name string,
	readFromBufSize uint,
) common.Conn {
	return &grpcConn{
		stream:          &bidiServerStream{stream: stream},
		closed:          false,
		buf:             []byte{},
		offset:          0,
		readFromBufSize: readFromBufSize,
		name:            name,
	}
}

// Write sends data through the gRPC stream.
// It implements the io.Writer interface for the connection.
// TODO: Split data if too large for a single gRPC message.
func (conn *grpcConn) Write(src []byte) (int, error) {
	if conn.closed {
		return 0, net.ErrClosed
	}

	if len(src) == 0 {
		return 0, nil
	}

	err := conn.stream.Send(src)

	if err != nil {
		return 0, err
	}
	return len(src), nil
}

// ReadFrom reads data from an io.Reader and sends it through the gRPC stream.
// It implements the io.ReaderFrom interface for efficient data transfer.
// Data is read in chunks and sent as separate gRPC messages.
func (conn *grpcConn) ReadFrom(r io.Reader) (int64, error) {
	if conn.closed {
		return 0, net.ErrClosed
	}

	buf := make([]byte, conn.readFromBufSize)
	var m int64

	for {
		n, err := r.Read(buf)
		m += int64(n)
		if err != nil && !isNormalClosedErr(err) {
			return m, err
		}
		_, werr := conn.Write(buf[:n])
		if werr != nil {
			return m, fmt.Errorf("failed to write %d bytes: %w", n, werr)
		}
		if err != nil {
			return m, nil
		}
	}
}

// Reading
func (conn *grpcConn) Read(dst []byte) (int, error) {
	if conn.closed {
		return 0, net.ErrClosed
	}

	m := 0
	for {
		n, isFull := conn.fillSlice(dst[m:])
		m += n
		if isFull {
			return m, nil
		}
		// TODO: We should return here and recv more next call?
		// dst is not full, so we ran out of data in conn.buf.
		// Therefore, we need to get more data.
		src, err := conn.stream.Recv()
		if err != nil {
			return m, err
		}
		conn.buf = src
		conn.offset = 0
	}
}

func (conn *grpcConn) WriteTo(dst io.Writer) (int64, error) {
	if conn.closed {
		return 0, net.ErrClosed
	}

	var m int64
	n, err := dst.Write(conn.buf[conn.offset:])
	m += int64(n)
	if err != nil {
		return m, err
	}
	conn.offset = 0
	conn.buf = []byte{}

	for {
		src, err := conn.stream.Recv()
		if isNormalClosedErr(err) {
			return m, nil
		}
		if err != nil {
			return m, err
		}
		n, err = dst.Write(src)
		m += int64(n)
		if err != nil {
			return m, err
		}
	}
}

func isNormalClosedErr(err error) bool {
	return err == io.EOF || status.Code(err) == codes.Canceled
}

func (conn *grpcConn) fillSlice(dst []byte) (int, bool) {
	n := copy(dst, conn.buf[conn.offset:])
	conn.offset += n
	return n, n >= len(dst)
}

func (conn *grpcConn) Close() error {
	return conn.stream.Close()
}

func (conn *grpcConn) Name() string {
	return conn.name
}
