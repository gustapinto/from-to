package webhook

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gustapinto/from-to/internal/event"
)

type Publisher struct {
	url     string
	retries uint64
	headers map[string]string
	client  *http.Client
	logger  *slog.Logger
}

func NewPublisher(config Config) *Publisher {
	return &Publisher{
		url:     config.URL,
		retries: config.RetriesOrDefault(),
		headers: config.HeadersOrDefault(),
		client: &http.Client{
			Timeout: config.TimeoutSecondsOrDefault(),
		},
		logger: slog.With("publisher", "Webhook"),
	}
}

func (p *Publisher) Publish(e event.Event, payload []byte) error {
	body := bytes.NewBuffer(payload)

	req, err := http.NewRequest(http.MethodPost, p.url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	for header, value := range p.headers {
		req.Header.Add(header, value)
	}

	tries := uint64(0)
	for tries < p.retries {
		res, err := p.client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode == http.StatusNoContent {
			return nil
		}

		p.logger.Debug("Endpoint returned non 204 response, retrying")

		tries += 1
	}

	p.logger.Debug(
		"Row published",
		"id", e.ID,
		"url", p.url,
		"payload", string(payload),
	)

	return fmt.Errorf("maximum tries exceeded for event %d", e.ID)
}
