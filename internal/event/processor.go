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

func (p *Processor) ListenAndProcess() error {
	return p.listener.Listen(p.publishEventToAllChannels)
}

func (p *Processor) publishEventToAllChannels(e Event, channels []Channel) error {
	var wg sync.WaitGroup
	for _, channel := range channels {
		wg.Add(1)

		p.logger.Debug("Publishing to channel", "event", e, "channel", channel.Key)

		go func() {
			defer wg.Done()

			if err := p.publishEventOnChannel(e, channel); err != nil {
				p.logger.Error(
					"Failed to process event for channel",
					"event", e.ID,
					"channel", channel.Key,
					"error", err.Error(),
				)
				return
			}

			p.logger.Debug(
				"Event published for channel",
				"event", e.ID,
				"channel", channel.Key,
			)
		}()
	}

	wg.Wait()

	return nil
}

func (p *Processor) publishEventOnChannel(e Event, channel Channel) error {
	publisher, err := p.getPublisher(channel)
	if err != nil {
		return err
	}

	payload, err := p.getPayload(e, channel)
	if err != nil {
		return err
	}

	if err := publisher.Publish(e, payload); err != nil {
		return err
	}

	return nil
}

func (p *Processor) getPublisher(channel Channel) (Publisher, error) {
	publisher, exists := p.publishers[channel.To]
	if !exists {
		return nil, fmt.Errorf(
			"failed to publish event, publisher [%s] dont exists",
			channel.To)
	}

	return publisher, nil
}

func (p *Processor) getPayload(e Event, channel Channel) ([]byte, error) {
	mapper, exists := p.mappers[channel.Mapper]
	if exists {
		return mapper.Map(e)
	}

	return json.Marshal(e)
}
