package broadcast

import (
	"context"

	"github.com/isacskoglund/rotox/internal/common"
)

type subscription[T any] struct {
	ctx     context.Context
	ch      <-chan T
	onClose func()
}

func newSubscription[T any](ctx context.Context, ch <-chan T, onClose func()) common.Subscription[T] {
	return &subscription[T]{
		ctx:     ctx,
		ch:      ch,
		onClose: onClose,
	}
}

func (s *subscription[T]) Receive() (T, error) {
	select {
	case <-s.ctx.Done():
		return *new(T), s.ctx.Err()
	case t := <-s.ch:
		return t, nil
	}
}

func (s *subscription[T]) Close() {
	s.onClose()
}
