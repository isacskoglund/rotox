package common

import "context"

// Subscriber provides the ability to subscribe to events of type T.
type Subscriber[T any] interface {
	// Subscribe creates a new subscription for receiving events.
	// The subscription remains active until the context is canceled or Close is called.
	Subscribe(context.Context) (Subscription[T], error)
}

// Subscription represents an active subscription to events of type T.
type Subscription[T any] interface {
	// Receive blocks until an event is available or an error occurs.
	// Returns an error if the subscription is closed or encounters an error.
	Receive() (T, error)

	// Close terminates the subscription and releases associated resources.
	Close()
}

// Publisher provides the ability to publish events of type T to subscribers.
type Publisher[T any] interface {
	// Publish sends an event to all active subscribers.
	// Returns an error if publishing fails.
	Publish(event T) error
}
