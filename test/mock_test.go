package integration_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockDialer struct {
	mock.Mock
}

func (m *mockDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	args := m.Called(ctx, network, address)
	return args.Get(0).(net.Conn), args.Error(1)
}

type mockConn struct {
	closeOnce sync.Once
	closeCh   chan struct{}
	readCh    chan byte
	writeCh   chan byte
}

func newMockConn(size int) *mockConn {
	return &mockConn{
		closeCh: make(chan struct{}),
		readCh:  make(chan byte, size),
		writeCh: make(chan byte, size),
	}
}

func (m *mockConn) toRead(src []byte) {
	for _, b := range src {
		m.readCh <- b
	}
}

func (m *mockConn) fromWrite(timeout time.Duration) []byte {
	result := bytes.NewBuffer([]byte{})
	timer := time.After(timeout)

	for {
		select {
		case written := <-m.writeCh:
			result.WriteByte(written)
		case <-timer:
			return result.Bytes()
		}
	}
}

func (m *mockConn) Close() error {
	m.closeOnce.Do(
		func() {
			close(m.closeCh)
		},
	)
	return nil
}

func (m *mockConn) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}
	select {
	case <-m.closeCh:
		return 0, io.EOF
	case b := <-m.readCh:
		dst[0] = b
		n := 1
	loop:
		for n < len(dst) && n < 10 {
			select {
			case b = <-m.readCh:
				dst[n] = b
				n++
			default:
				break loop
			}
		}
		return n, nil
	}
}

func (m *mockConn) Write(src []byte) (int, error) {
	n := 0
	for {
		if n == len(src) {
			return n, nil
		}
		select {
		case <-m.closeCh:
			return n, io.ErrClosedPipe
		case m.writeCh <- src[n]:
			n++
		}
	}
}

func (m *mockConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (m *mockConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (m *mockConn) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}
