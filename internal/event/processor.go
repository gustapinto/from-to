package event

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

type Processor struct {
	listener  Listener
	publisher Publisher
	logger    *slog.Logger
}

func NewProcessor(listener Listener, publisher Publisher) *Processor {
	return &Processor{
		listener:  listener,
		publisher: publisher,
		logger:    slog.Default(),
	}
}

func (s *Processor) ListenAndProcess() error {
	err := s.listener.Listen(func(event Event) error {
		var (
			payload []byte
			err     error
		)

		if event.Metadata.Mapper != nil {
			payload, err = event.Metadata.Mapper.Map(event)
			if err != nil {
				return fmt.Errorf("failed to map event payload, eventId=%d", event.ID)
			}
		} else {
			payload, err = json.Marshal(event)
			if err != nil {
				return fmt.Errorf("failed to process event payload, eventId=%d", event.ID)
			}
		}

		if err := s.publisher.Publish(event, payload); err != nil {
			return fmt.Errorf("failed to process event payload, eventId=%d", event.ID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
