package events

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// WatermillEventManager implements EventManager using Watermill
type WatermillEventManager struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	logger     *zerolog.Logger
}

// NewWatermillEventManager creates a new WatermillEventManager
func NewWatermillEventManager(logger *zerolog.Logger) (*WatermillEventManager, error) {
	// Initialize Watermill publisher/subscriber
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	return &WatermillEventManager{
		publisher:  pubSub,
		subscriber: pubSub,
		logger:     logger,
	}, nil
}

// Subscribe implements EventManager.Subscribe
func (m *WatermillEventManager) Subscribe(ctx context.Context, pageID string) (<-chan UIEvent, error) {
	topic := fmt.Sprintf("ui-updates.%s", pageID)
	messages, err := m.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	events := make(chan UIEvent)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer close(events)
		for msg := range messages {
			event, err := FromJSON(msg.Payload)
			if err != nil {
				m.logger.Error().Err(err).Msg("Failed to parse event message")
				msg.Nack()
				continue
			}
			select {
			case events <- event:
				msg.Ack()
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	// We don't need to wait for the errgroup to finish here since we want to return the channel immediately
	// The goroutine will clean itself up when the context is cancelled or messages channel is closed
	return events, nil
}

// Publish implements EventManager.Publish
func (m *WatermillEventManager) Publish(pageID string, event UIEvent) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	payload, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := message.NewMessage(
		watermill.NewUUID(),
		payload,
	)
	msg.Metadata = map[string]string{
		"event-type": event.Type,
	}

	topic := fmt.Sprintf("ui-updates.%s", pageID)
	if err := m.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Close implements EventManager.Close
func (m *WatermillEventManager) Close() error {
	var errs []error

	if err := m.publisher.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close publisher: %w", err))
	}

	if err := m.subscriber.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close subscriber: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close event manager: %v", errs)
	}

	return nil
}
