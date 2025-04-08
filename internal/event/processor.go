package event

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

type Processor struct {
	listener   Listener
	publishers map[string]Publisher
	mappers    map[string]Mapper
	channels   map[string]Channel
	logger     *slog.Logger
}

func NewProcessor(
	listener Listener,
	publishers map[string]Publisher,
	mappers map[string]Mapper,
	channels map[string]Channel,
) *Processor {
	return &Processor{
		listener:   listener,
		publishers: publishers,
		mappers:    mappers,
		channels:   channels,
		logger:     slog.Default(),
	}
}

func (s *Processor) ListenAndProcess() error {
	return s.listener.Listen(s.publishEventToAllChannels)
}

func (s *Processor) publishEventToAllChannels(e Event) error {
	var wg sync.WaitGroup
	for _, channel := range e.Channels {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := s.publishEventOnChannel(e, channel); err != nil {
				s.logger.Error(
					"Failed to process event for channel",
					"event", e.ID,
					"channel", channel.Key,
					"error", err.Error(),
				)
				return
			}

			s.logger.Debug(
				"Event processed for channel",
				"event", e.ID,
				"channel", channel.Key,
			)
		}()
	}

	wg.Wait()

	return nil
}

func (s *Processor) publishEventOnChannel(e Event, channel Channel) error {
	publisher, err := s.getPublisher(channel)
	if err != nil {
		return err
	}

	payload, err := s.getPayload(e, channel)
	if err != nil {
		return err
	}

	if err := publisher.Publish(e, payload, channel.To); err != nil {
		return err
	}

	return nil
}

func (s *Processor) getPublisher(channel Channel) (Publisher, error) {
	publisher, exists := s.publishers[channel.Output]
	if !exists {
		return nil, fmt.Errorf(
			"failed to publish event, publisher [%s] dont exists",
			channel.Output)
	}

	return publisher, nil
}

func (s *Processor) getPayload(e Event, channel Channel) ([]byte, error) {
	mapper, exists := s.mappers[channel.Mapper]
	if exists {
		return mapper.Map(e)
	}

	return json.Marshal(e)
}
