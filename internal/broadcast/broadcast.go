// Package broadcast provides a generic pub-sub broadcasting system.
//
// This package implements a type-safe broadcaster that allows multiple
// subscribers to receive events of type T. It's used throughout rotox
// for distributing telemetry events and other system notifications.
//
// The broadcaster is designed to be non-blocking for publishers and
// provides automatic cleanup of subscribers when they disconnect.
package broadcast

import (
	"context"
	"fmt"
	"sync"

	"github.com/isacskoglund/rotox/internal/common"
)

// Broadcaster provides a generic pub-sub system for events of type T.
// It manages multiple subscribers and broadcasts events to all active subscribers.
type Broadcaster[T any] struct {
	isRunning   sync.Mutex  // Protects broadcaster state
	publish     chan T      // Channel for incoming events to broadcast
	subscribe   chan chan T // Channel for new subscriber registrations
	unsubscribe chan chan T // Channel for subscriber cleanup

	subscriberChannelSize int // Buffer size for subscriber channels
}

// NewBroadcaster creates a new broadcaster instance for events of type T.
// TODO: Move these magic values to configuration.
func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		isRunning:             sync.Mutex{},
		publish:               make(chan T, 10),
		subscribe:             make(chan chan T, 10),
		unsubscribe:           make(chan chan T, 10),
		subscriberChannelSize: 10,
	}
}

func (p *Broadcaster[T]) Subscribe(ctx context.Context) (common.Subscription[T], error) {
	if err := p.ensureIsRunning(); err != nil {
		return nil, err
	}
	ch := make(chan T, p.subscriberChannelSize)
	p.subscribe <- ch
	onClose := func() {
		p.unsubscribe <- ch
	}

	return newSubscription(ctx, ch, onClose), nil
}

func (p *Broadcaster[T]) Publish(value T) error {
	if err := p.ensureIsRunning(); err != nil {
		return err
	}
	p.publish <- value
	return nil
}

func (b *Broadcaster[T]) ensureIsRunning() error {
	if !b.isRunning.TryLock() {
		return nil
	}
	defer b.isRunning.Unlock()
	return fmt.Errorf("broadcaster is not started")
}

func (p *Broadcaster[T]) Start(ctx context.Context) error {
	ok := p.isRunning.TryLock()
	if !ok {
		return fmt.Errorf("broadcaster is already started")
	}

	go func() {
		defer p.isRunning.Unlock()

		subscribers := make(map[chan<- T]struct{})

		publish := func(value T) {
			for ch := range subscribers {
				select {
				case ch <- value:
				default:
					// TODO: Use a real logger
					fmt.Println("failed to send to subscriber")
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				for ch := range subscribers {
					close(ch)
				}
			case value := <-p.publish:
				publish(value)
			case ch := <-p.subscribe:
				subscribers[ch] = struct{}{}
			case ch := <-p.unsubscribe:
				_, ok := subscribers[ch]
				if !ok {
					// TODO: Use a real logger
					fmt.Println("tried to unsubscribe a subscriber that does not exist")
				}
				delete(subscribers, ch)
				close(ch)
			}
		}
	}()
	return nil
}
