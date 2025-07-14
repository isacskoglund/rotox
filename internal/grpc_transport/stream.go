package grpc_transport

import (
	"fmt"

	forward_pb "github.com/isacskoglund/rotox/gen/go/forward/v1"
	"google.golang.org/grpc"
)

// bidiStream provides a simplified interface for bidirectional streaming
// over gRPC. It abstracts the protobuf message wrapping/unwrapping.
type bidiStream interface {
	Send([]byte) error     // Send raw bytes through the stream
	Recv() ([]byte, error) // Receive raw bytes from the stream
	Close() error          // Close the stream
}

// bidiClientStream wraps grpc.BidiStreamingClient to implement bidiStream.
// It handles the client-side protobuf message marshaling for data transfer.
type bidiClientStream struct {
	stream grpc.BidiStreamingClient[forward_pb.ForwardRequest, forward_pb.ForwardResponse]
}

// Send wraps the provided bytes in a TransferRequest and sends it through the stream.
func (stream *bidiClientStream) Send(src []byte) error {
	return stream.stream.Send(
		&forward_pb.ForwardRequest{
			Request: &forward_pb.ForwardRequest_TransferRequest{
				TransferRequest: &forward_pb.TransferRequest{
					Data: src,
				},
			},
		},
	)
}

// Recv receives a message from the stream and extracts the data bytes.
// It expects TransferResponse messages and returns an error if the
// response is not of the expected type.
func (stream *bidiClientStream) Recv() ([]byte, error) {
	msg, err := stream.stream.Recv()
	if err != nil {
		return nil, err
	}
	forwardResponse := msg.GetTransferResponse()
	if forwardResponse == nil {
		return nil, fmt.Errorf("transfer response was nil")
	}
	return forwardResponse.Data, nil
}

// Close closes the client-side stream by calling CloseSend.
func (stream *bidiClientStream) Close() error {
	return stream.stream.CloseSend()
}

// bidiServerStream wraps grpc.BidiStreamingServer to implement bidiStream.
// It handles the server-side protobuf message marshaling for data transfer.
type bidiServerStream struct {
	stream grpc.BidiStreamingServer[forward_pb.ForwardRequest, forward_pb.ForwardResponse]
}

// Send wraps the provided bytes in a TransferResponse and sends it through the stream.
func (stream *bidiServerStream) Send(src []byte) error {
	return stream.stream.Send(
		&forward_pb.ForwardResponse{
			Response: &forward_pb.ForwardResponse_TransferResponse{
				TransferResponse: &forward_pb.TransferResponse{
					Data: src,
				},
			},
		},
	)
}

// Recv receives a message from the stream and extracts the data bytes.
// It expects TransferRequest messages and returns an error if the
// request is not of the expected type.
func (stream *bidiServerStream) Recv() ([]byte, error) {
	msg, err := stream.stream.Recv()
	if err != nil {
		return nil, err
	}
	transferRequest := msg.GetTransferRequest()
	if transferRequest == nil {
		return nil, fmt.Errorf("transfer request was nil")
	}
	return transferRequest.Data, nil
}

func (stream *bidiServerStream) Close() error {
	return nil
}
