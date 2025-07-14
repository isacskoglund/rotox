package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024

func TestConnectWithSingleProbe(t *testing.T) {

	logLevel := slog.LevelDebug

	// 1. Spin up hub and probe in separate goroutines.
	target := "www.example.com"
	targetConn := newMockConn(1024)
	httpLis := bufconn.Listen(bufSize)
	defer httpLis.Close()
	{
		targetDialer := &mockDialer{}
		grpcLis := bufconn.Listen(bufSize)
		defer grpcLis.Close()
		targetDialer.On("DialContext", mock.Anything, "tcp", target).Once().Return(targetConn, nil)
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
		serveProbe(grpcLis, logger.With("logger", "probe"), targetDialer)
		serveHub(httpLis, logger.With("logger", "hub"), []*bufconn.Listener{grpcLis})
	}

	// 2. Establish connection
	httpConn, err := httpLis.DialContext(context.Background())
	{
		connectRequest := http.Request{
			Method: "CONNECT",
			Host:   target,
			URL: &url.URL{
				Opaque: target,
			},
		}
		assert.NoError(t, err, "dial httpLis")
		err = connectRequest.Write(httpConn)
		assert.NoError(t, err, "write connectRequest to httpConn")
		res, err := http.ReadResponse(
			bufio.NewReader(httpConn),
			nil,
		)
		assert.NoError(t, err, "read connect response")
		assert.Equal(t, res.StatusCode, http.StatusOK, "connect response status code")
	}

	// 3. Send data
	{
		dataToSend := []byte("some-random-content-being-sent")
		_, err = httpConn.Write(dataToSend)
		assert.NoError(t, err)
		written := targetConn.fromWrite(time.Second)
		assert.Equal(t, dataToSend, written)
	}

	// 4. Receive data
	{
		dataToReceive := []byte("some-random-content-being-received")
		targetConn.toRead(dataToReceive)
		received := bytes.NewBuffer([]byte{})
		httpConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, err := io.Copy(received, httpConn)
		assert.EqualError(t, err, "i/o timeout")
		assert.Equal(t, dataToReceive, received.Bytes())
		httpConn.SetReadDeadline(time.Time{})
	}

	// 5. Close the connection from the target side
	{
		targetConn.Close()
		received := make([]byte, 10)
		n, err := httpConn.Read(received)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, 0, n)
	}
}
