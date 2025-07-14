package grpc_transport_test

import (
	"context"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"github.com/isacskoglund/rotox/internal/common"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockBidiStream[Snd any, Rcv any] struct {
	mock.Mock
}

func newMockForwardBidiStreamingClient() *mockBidiStream[*forward_pb.ForwardRequest, *forward_pb.ForwardResponse] {
	return &mockBidiStream[*forward_pb.ForwardRequest, *forward_pb.ForwardResponse]{}
}

func newMockForwardBidiStreamingServer() *mockBidiStream[*forward_pb.ForwardResponse, *forward_pb.ForwardRequest] {
	return &mockBidiStream[*forward_pb.ForwardResponse, *forward_pb.ForwardRequest]{}
}

func (m *mockBidiStream[Snd, Rcv]) onSend(
	req Snd,
	returnErr error,
) *mock.Call {
	return m.Mock.On("Send", req).Return(returnErr)
}

func (m *mockBidiStream[Snd, Rcv]) onRecv(
	returnRcv Rcv,
	returnError error,
) *mock.Call {
	return m.Mock.On("Recv").Return(returnRcv, returnError)
}

func (m *mockBidiStream[Snd, Rcv]) Send(req Snd) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *mockBidiStream[Snd, Rcv]) Recv() (Rcv, error) {
	args := m.Called()
	return args.Get(0).(Rcv), args.Error(1)
}

func (m *mockBidiStream[Snd, Rcv]) CloseSend() error {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *mockBidiStream[Snd, Rcv]) Header() (metadata.MD, error) {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) Trailer() metadata.MD {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) RecvMsg(msg any) error {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) SendMsg(msg any) error {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) SetHeader(metadata.MD) error {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) SendHeader(metadata.MD) error {
	panic("not implemented")
}

func (m *mockBidiStream[Snd, Rcv]) SetTrailer(metadata.MD) {
	panic("not implemented")
}

type mockForwardServiceClient struct {
	mock.Mock
}

func newMockForwardServiceClient() *mockForwardServiceClient {
	return &mockForwardServiceClient{}
}

func (m *mockForwardServiceClient) onForward(
	opts []grpc.CallOption,
	returnClient forward_pb.ForwardService_ForwardClient,
	returnError error,
) *mock.Call {
	return m.Mock.On("Forward", mock.Anything, opts).Return(returnClient, returnError)
}

func (m *mockForwardServiceClient) Forward(ctx context.Context, opts ...grpc.CallOption) (forward_pb.ForwardService_ForwardClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(forward_pb.ForwardService_ForwardClient), args.Error(1)
}

type mockForwarder struct {
	mock.Mock
}

func newMockForwarder() *mockForwarder {
	return &mockForwarder{}
}

func (m *mockForwarder) Forward(
	ctx context.Context,
	targetAddress string,
	accept func() (common.Conn, error),
) error {
	args := m.Called(ctx, targetAddress, accept)
	return args.Error(0)
}
