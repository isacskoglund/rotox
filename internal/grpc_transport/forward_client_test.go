package grpc_transport_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/isacskoglund/rotox/internal/fault"
	"github.com/isacskoglund/rotox/internal/grpc_transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestForwardClient_ReadAndWriteTo(t *testing.T) {
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
		run  func(
			t *testing.T,
			c *Case,
			conn common.Conn,
		)
	}

	testRead := connTest{
		name: "Read",
		run: func(
			t *testing.T,
			c *Case,
			conn common.Conn,
		) {
			// Arrange
			expectedResult := []byte(strings.Join(c.messageParts, ""))
			// allocate more space than needed,
			// this allows us to detect if Write writes more that it should
			padding := make([]byte, 100)
			buf := make([]byte, len(expectedResult)+len(padding))

			// Act
			n, err := conn.Read(buf)

			// Assert
			assert.Equal(t, err, io.EOF)
			assert.Equal(t, len(expectedResult), n)
			assert.Equal(t, string(expectedResult), string(buf[:len(expectedResult)]))
			assert.Equal(t, padding, buf[len(expectedResult):]) // should be all zeros
		},
	}

	testWriteTo := connTest{
		name: "WriteTo",
		run: func(
			t *testing.T,
			c *Case,
			conn common.Conn,
		) {
			// Arrange
			writerTo, ok := any(conn).(io.WriterTo)
			if !ok {
				t.Fatal("conn does not implement io.WriterTo")
			}
			var buf strings.Builder
			expected := strings.Join(c.messageParts, "")

			// Act
			n, err := writerTo.WriteTo(&buf)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, int64(len(expected)), n)
			assert.Equal(t, expected, buf.String())
		},
	}

	for _, c := range cases {
		for _, connTest := range []connTest{testRead, testWriteTo} {
			t.Run(
				fmt.Sprintf("%s/%s", c.name, connTest.name),
				func(t *testing.T) {
					// Arrange: Init
					ctx := context.Background()
					mockClient := newMockForwardServiceClient()
					mockStream := newMockForwardBidiStreamingClient()
					client := grpc_transport.NewForwardClient(mockClient)

					// Arrange: On
					mockClient.onForward(nil, mockStream, nil).Once()
					mockStream.onSend(
						&forward_pb.ForwardRequest{
							Request: &forward_pb.ForwardRequest_DialRequest{
								DialRequest: &forward_pb.DialRequest{
									Destination: c.target,
								},
							},
						},
						nil,
					).Once()
					mockStream.onRecv(
						&forward_pb.ForwardResponse{
							Response: &forward_pb.ForwardResponse_DialResponse{
								DialResponse: &forward_pb.DialResponse{
									Code: forward_pb.DialResponse_CODE_UNSPECIFIED,
								},
							},
						},
						nil,
					).Once()
					for _, part := range c.messageParts {
						mockStream.onRecv(
							&forward_pb.ForwardResponse{
								Response: &forward_pb.ForwardResponse_TransferResponse{
									TransferResponse: &forward_pb.TransferResponse{
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

					// Act
					conn, err := client.Dial(ctx, c.target)

					// Assert
					assert.NoError(t, err)

					connTest.run(t, &c, conn)

					mockStream.Mock.AssertExpectations(t)
					mockClient.Mock.AssertExpectations(t)
				},
			)
		}
	}
}

func TestForwardClient_WriteAndReadFrom(t *testing.T) {
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
		run  func(
			t *testing.T,
			c *Case,
			conn common.Conn,
			sentMessages *[][]byte,
		)
	}

	testWrite := connTest{
		name: "Write",
		run: func(
			t *testing.T,
			c *Case,
			conn common.Conn,
			sentMessages *[][]byte,
		) {
			// Arrange
			expected := strings.Join(c.messageParts, "")

			// Act
			for _, part := range c.messageParts {
				n, err := conn.Write([]byte(part))
				assert.NoError(t, err)
				assert.Equal(t, len(part), n)
			}

			// Assert
			assert.Equal(t, expected, string(bytes.Join(*sentMessages, []byte{})))
		},
	}

	testReadFrom := connTest{
		name: "ReadFrom",
		run: func(
			t *testing.T,
			c *Case,
			conn common.Conn,
			sentMessages *[][]byte,
		) {
			// Arrange
			readerFrom := conn.(io.ReaderFrom)
			reader := strings.NewReader(strings.Join(c.messageParts, ""))
			expected := strings.Join(c.messageParts, "")

			// Act
			n, err := readerFrom.ReadFrom(reader)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, int64(len(expected)), n)
			assert.Equal(t, expected, string(bytes.Join(*sentMessages, []byte{})))
		},
	}

	for _, c := range cases {
		for _, connTest := range []connTest{testWrite, testReadFrom} {
			t.Run(
				fmt.Sprintf("%s/%s", c.name, connTest.name),
				func(t *testing.T) {
					ctx := context.Background()
					mockClient := newMockForwardServiceClient()
					mockStream := newMockForwardBidiStreamingClient()
					client := grpc_transport.NewForwardClient(mockClient)
					var sentMessages [][]byte

					// Arrange: On
					mockClient.onForward(nil, mockStream, nil).Once()
					mockStream.onSend(
						&forward_pb.ForwardRequest{
							Request: &forward_pb.ForwardRequest_DialRequest{
								DialRequest: &forward_pb.DialRequest{
									Destination: c.target,
								},
							},
						},
						nil,
					).Once()
					mockStream.onRecv(
						&forward_pb.ForwardResponse{
							Response: &forward_pb.ForwardResponse_DialResponse{
								DialResponse: &forward_pb.DialResponse{
									Code: forward_pb.DialResponse_CODE_UNSPECIFIED,
								},
							},
						},
						nil,
					).Once()
					mockStream.On("Send", mock.Anything).Run(func(args mock.Arguments) {
						req := args.Get(0).(*forward_pb.ForwardRequest)
						transfer := req.GetTransferRequest()
						if transfer == nil {
							t.Fatal("mockStream.Send received ForwardRequest where TransferRequest is nil")
						}
						sentMessages = append(sentMessages, transfer.Data)
					}).Return(nil)

					// Act
					conn, err := client.Dial(ctx, c.target)
					assert.NoError(t, err)

					connTest.run(t, &c, conn, &sentMessages)

					mockStream.Mock.AssertExpectations(t)
					mockClient.Mock.AssertExpectations(t)
				},
			)
		}
	}
}

func TestForwardClient_Dial_ErrorCodes(t *testing.T) {
	type Case struct {
		name          string
		target        string
		dialResp      *forward_pb.ForwardResponse
		recvErr       error
		expectErr     bool
		expectErrCode any
	}

	cases := []Case{
		{
			name:   "success (unspecified code)",
			target: "ok.com:80",
			dialResp: &forward_pb.ForwardResponse{
				Response: &forward_pb.ForwardResponse_DialResponse{
					DialResponse: &forward_pb.DialResponse{
						Code: forward_pb.DialResponse_CODE_UNSPECIFIED,
					},
				},
			},
			recvErr:       nil,
			expectErr:     false,
			expectErrCode: "",
		},
		{
			name:   "failed to resolve host",
			target: "failresolve.com:80",
			dialResp: &forward_pb.ForwardResponse{
				Response: &forward_pb.ForwardResponse_DialResponse{
					DialResponse: &forward_pb.DialResponse{
						Code: forward_pb.DialResponse_CODE_FAILED_TO_RESOLVE_HOST,
					},
				},
			},
			recvErr:       nil,
			expectErr:     true,
			expectErrCode: common.ForwardFailedToResolveHost,
		},
		{
			name:   "host unreachable",
			target: "unreachable.com:80",
			dialResp: &forward_pb.ForwardResponse{
				Response: &forward_pb.ForwardResponse_DialResponse{
					DialResponse: &forward_pb.DialResponse{
						Code: forward_pb.DialResponse_CODE_HOST_UNREACHABLE,
					},
				},
			},
			recvErr:       nil,
			expectErr:     true,
			expectErrCode: common.ForwardHostUnreachable,
		},
		{
			name:          "grpc error (Unavailable)",
			target:        "grpcerror.com:80",
			dialResp:      nil,
			recvErr:       status.Error(codes.Unavailable, "unavailable"),
			expectErr:     true,
			expectErrCode: common.ForwardUnknown,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Arrange: Init
			ctx := context.Background()
			mockClient := newMockForwardServiceClient()
			mockStream := newMockForwardBidiStreamingClient()
			client := grpc_transport.NewForwardClient(mockClient)

			// Arrange: On
			mockClient.onForward(nil, mockStream, nil).Once()
			mockStream.On("Send", mock.Anything).Return(nil)
			if c.dialResp != nil {
				mockStream.On("Recv").Once().Return(c.dialResp, nil)
			} else {
				mockStream.On("Recv").Once().Return(&forward_pb.ForwardResponse{}, c.recvErr)
			}

			// Act
			conn, err := client.Dial(ctx, c.target)

			// Assert
			if c.expectErr {
				assert.Nil(t, conn)
				assert.Error(t, err)
				assert.Equal(t, c.expectErrCode, fault.RawCode(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
			}

			mockStream.Mock.AssertExpectations(t)
			mockClient.Mock.AssertExpectations(t)
		})
	}
}
