package grpc_transport_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/grpc_transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestForwardServer_ReadAndWriteTo(t *testing.T) {
	type Case struct {
		name         string
		target       string
		messageParts []string
	}

	cases := []Case{
		{
			name:         "single message",
			target:       "example.com:80",
			messageParts: []string{"hello world"},
		},
		{
			name:         "multiple messages",
			target:       "example.com:443",
			messageParts: []string{"hello", " ", "world", "!"},
		},
		{
			name:         "empty messages",
			target:       "localhost:8000",
			messageParts: []string{"", "hello", "", "world", ""},
		},
		{
			name:   "long single message",
			target: "longhost.com:9090",
			messageParts: []string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit." +
					" Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			},
		},
		{
			name:   "long multiple messages",
			target: "multiparts.org:7070",
			messageParts: []string{
				"The quick brown fox ",
				"jumps over the lazy dog. ",
				"This sentence is used to test fonts ",
				"and keyboard layouts across systems.",
			},
		},
		{
			name:   "long messages with empty parts",
			target: "data.net:6000",
			messageParts: []string{
				"",
				"This is a long message with ",
				"intermittent empty parts to ",
				"simulate possible real-world scenarios ",
				"where data chunks may be missing.",
				"",
			},
		},
		{
			name:   "unicode and symbols",
			target: "symbols.io:5050",
			messageParts: []string{
				"Unicode test: 測試, テスト, اختبار, тест",
				" — special chars: !@#$%^&*()_+{}|:\"<>?`~[];',./",
			},
		},
		{
			name:   "very long sentence broken in parts",
			target: "fragmented.net:3000",
			messageParts: []string{
				"This is part one of a very long message that is meant to be sent in chunks, ",
				"each of which represents a logical break in the overall structure, ",
				"perhaps for buffering, streaming, or networking reasons where ",
				"breaking up a message improves efficiency or matches protocol design.",
			},
		},
	}

	type connTest struct {
		name string
		read func(*testing.T, common.Conn, []string)
	}

	testRead := connTest{
		name: "Read",
		read: func(t *testing.T, conn common.Conn, messageParts []string) {
			// Arrange
			expected := strings.Join(messageParts, "")
			padding := make([]byte, 100)
			buf := make([]byte, len(expected)+len(padding))

			// Act
			n, err := conn.Read(buf)

			// Assert
			assert.Equal(t, err, io.EOF)
			assert.Equal(t, len(expected), n)
			assert.Equal(t, expected, string(buf[:len(expected)]))
			assert.Equal(t, padding, buf[len(expected):])
		},
	}

	testWriteTo := connTest{
		name: "WriteTo",
		read: func(t *testing.T, conn common.Conn, messageParts []string) {
			writerTo, ok := conn.(io.WriterTo)
			if !ok {
				t.Fatal("conn is not an io.WriterTo")
			}
			expected := []byte(strings.Join(messageParts, ""))
			var out bytes.Buffer
			n, err := writerTo.WriteTo(&out)
			assert.NoError(t, err)
			assert.Equal(t, int64(len(expected)), n)
			assert.Equal(t, expected, out.Bytes())
		},
	}

	for _, c := range cases {
		for _, connTest := range []connTest{testRead, testWriteTo} {
			t.Run(
				c.name+"/"+connTest.name,
				func(t *testing.T) {
					// Arrange: Init
					ctx := context.Background()
					mockForwarder := newMockForwarder()
					mockStream := newMockForwardBidiStreamingServer()
					logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
					server := grpc_transport.NewForwardServer(logger, mockForwarder)

					// Arrange: On
					mockStream.On("Context").Maybe().Return(ctx)
					mockStream.On("Recv").Once().Return(&forward_pb.ForwardRequest{
						Request: &forward_pb.ForwardRequest_DialRequest{
							DialRequest: &forward_pb.DialRequest{
								Destination: c.target,
							},
						},
					}, nil)
					mockStream.On("Send", mock.Anything).Once().Run(
						func(args mock.Arguments) {
							resp := args.Get(0).(*forward_pb.ForwardResponse)
							dialResp := resp.GetDialResponse()
							if assert.NotNil(t, dialResp) {
								assert.Equal(t, dialResp.Code, forward_pb.DialResponse_CODE_UNSPECIFIED)
							}
						},
					).Return(nil).Once()
					for _, part := range c.messageParts {
						mockStream.onRecv(
							&forward_pb.ForwardRequest{
								Request: &forward_pb.ForwardRequest_TransferRequest{
									TransferRequest: &forward_pb.TransferRequest{
										Data: []byte(part),
									},
								},
							},
							nil,
						).Once()
					}
					mockStream.onRecv(
						nil,
						io.EOF,
					).Once()
					mockForwarder.On("Forward", mock.Anything, c.target, mock.Anything).Run(func(args mock.Arguments) {
						accept := args.Get(2).(func() (common.Conn, error))
						conn, err := accept()
						assert.NoError(t, err)
						connTest.read(t, conn, c.messageParts)
					}).Return(nil).Once()

					// Act
					err := server.Forward(mockStream)

					// Assert
					assert.NoError(t, err)
					mockStream.AssertExpectations(t)
					mockForwarder.AssertExpectations(t)
				},
			)
		}
	}
}

func TestForwardServer_WriteAndReadFrom(t *testing.T) {
	type Case struct {
		name         string
		target       string
		messageParts []string
	}

	cases := []Case{
		{
			name:         "single message",
			target:       "example.com:80",
			messageParts: []string{"hello world"},
		},
		{
			name:         "multiple messages",
			target:       "example.com:443",
			messageParts: []string{"hello", " ", "world", "!"},
		},
		{
			name:         "empty messages",
			target:       "localhost:8000",
			messageParts: []string{"", "hello", "", "world", ""},
		},
		{
			name:   "long single message",
			target: "longhost.com:9090",
			messageParts: []string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit." +
					" Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			},
		},
		{
			name:   "long multiple messages",
			target: "multiparts.org:7070",
			messageParts: []string{
				"The quick brown fox ",
				"jumps over the lazy dog. ",
				"This sentence is used to test fonts ",
				"and keyboard layouts across systems.",
			},
		},
		{
			name:   "long messages with empty parts",
			target: "data.net:6000",
			messageParts: []string{
				"",
				"This is a long message with ",
				"intermittent empty parts to ",
				"simulate possible real-world scenarios ",
				"where data chunks may be missing.",
				"",
			},
		},
		{
			name:   "unicode and symbols",
			target: "symbols.io:5050",
			messageParts: []string{
				"Unicode test: 測試, テスト, اختبار, тест",
				" — special chars: !@#$%^&*()_+{}|:\"<>?`~[];',./",
			},
		},
		{
			name:   "very long sentence broken in parts",
			target: "fragmented.net:3000",
			messageParts: []string{
				"This is part one of a very long message that is meant to be sent in chunks, ",
				"each of which represents a logical break in the overall structure, ",
				"perhaps for buffering, streaming, or networking reasons where ",
				"breaking up a message improves efficiency or matches protocol design.",
			},
		},
	}

	type connTest struct {
		name  string
		write func(*testing.T, common.Conn, []string)
	}

	testWrite := connTest{
		name: "Write",
		write: func(t *testing.T, conn common.Conn, messageParts []string) {
			for _, part := range messageParts {
				n, err := conn.Write([]byte(part))
				assert.NoError(t, err, "error when writing to conn")
				assert.Equal(t, len([]byte(part)), n, "number of bytes written to conn does not the length of []byte(part)")
			}
		},
	}

	testReadFrom := connTest{
		name: "ReadFrom",
		write: func(t *testing.T, conn common.Conn, messageParts []string) {
			readerFrom, ok := conn.(io.ReaderFrom)
			if !ok {
				t.Fatal("conn does not implement io.ReaderFrom")
			}
			content := []byte(strings.Join(messageParts, ""))
			reader := bytes.NewReader(content)
			n, err := readerFrom.ReadFrom(reader)
			assert.NoError(t, err)
			assert.Equal(t, len(content), int(n), "n returned by ReadFrom does not match len(content)")
		},
	}

	for _, c := range cases {
		for _, connTest := range []connTest{testWrite, testReadFrom} {
			t.Run(
				c.name+"/"+connTest.name,
				func(t *testing.T) {
					// Arrange: Init
					ctx := context.Background()
					mockForwarder := newMockForwarder()
					mockStream := newMockForwardBidiStreamingServer()
					logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
					server := grpc_transport.NewForwardServer(logger, mockForwarder)
					var sentMessages [][]byte

					// Arrange: On
					mockStream.On("Context").Maybe().Return(ctx)
					mockStream.On("Recv").Once().Return(&forward_pb.ForwardRequest{
						Request: &forward_pb.ForwardRequest_DialRequest{
							DialRequest: &forward_pb.DialRequest{
								Destination: c.target,
							},
						},
					}, nil)
					mockStream.On("Send", mock.Anything).Once().Run(
						func(args mock.Arguments) {
							resp := args.Get(0).(*forward_pb.ForwardResponse)
							dialResp := resp.GetDialResponse()
							if assert.NotNil(t, dialResp) {
								assert.Equal(t, dialResp.Code, forward_pb.DialResponse_CODE_UNSPECIFIED)
							}
						},
					).Return(nil).Once()
					mockStream.On("Send", mock.Anything).Run(
						func(args mock.Arguments) {
							resp := args.Get(0).(*forward_pb.ForwardResponse)
							transferResp := resp.GetTransferResponse()
							if transferResp == nil {
								t.Fatal("Send received a ForwardResponse where TransferResponse was nil")
							}
							sentMessages = append(sentMessages, transferResp.Data)
						},
					).Return(nil)
					mockForwarder.On("Forward", mock.Anything, c.target, mock.Anything).Run(func(args mock.Arguments) {
						accept := args.Get(2).(func() (common.Conn, error))
						conn, err := accept()
						assert.NoError(t, err)
						connTest.write(t, conn, c.messageParts)
					}).Return(nil).Once()

					// Act
					err := server.Forward(mockStream)

					// Assert
					assert.NoError(t, err)
					assert.Equal(
						t,
						string(bytes.Join(sentMessages, []byte{})),
						strings.Join(c.messageParts, ""),
					)

					mockStream.AssertExpectations(t)
					mockForwarder.AssertExpectations(t)
				},
			)
		}
	}
}
