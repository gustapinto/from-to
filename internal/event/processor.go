package event

import (
	"encoding/json"
	"log/slog"
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
	err := s.listener.Listen(func(e Event) error {
		for _, channel := range e.Channels {
			publisher, exists := s.publishers[channel.Output]
			if !exists {
				s.logger.Error("Failed to publish event, publisher dont exists", "publisher", channel.Output)
				continue
			}

			var payload []byte
			var err error

			mapper, exists := s.mappers[channel.Mapper]
			if exists {
				payload, err = mapper.Map(e)
			} else {
				payload, err = json.Marshal(e)
			}

			if err != nil {
				s.logger.Error("Failed to parse payload from event", "error", err.Error())
				continue
			}

			if err := publisher.Publish(e, payload, channel.To); err != nil {
				s.logger.Error("Failed to publish event", "error", err.Error())
				continue
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
