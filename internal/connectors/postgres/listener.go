package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gustapinto/from-to/internal/event"
	_ "github.com/lib/pq"
)

type Listener struct {
	dsn                    string
	limit                  uint64
	waitSeconds            time.Duration
	timeout                time.Duration
	db                     *sql.DB
	logger                 *slog.Logger
	tableToChannelRelation map[string][]event.Channel
}

func NewListener(config Config, channels map[string]event.Channel) (*Listener, error) {
	listener := &Listener{
		dsn:         config.DSN,
		limit:       config.LimitOrDefault(),
		waitSeconds: config.PollSecondsOrDefault(),
		timeout:     config.TimeoutSecondsOrDefault(),
		logger:      slog.With("listener", "Postgres"),
	}

	if err := listener.connectToDatabase(config.DSN); err != nil {
		return nil, err
	}

	if err := listener.setupDatabaseSchema(config); err != nil {
		return nil, err
	}

	listener.setupTableToChannelRelation(channels)
	listener.logger.Info("Connector setup completed")

	return listener, nil
}

func (l *Listener) Listen(callback func(event.Event, []event.Channel) error) error {
	limit := uint64(50)

	for {
		events, err := l.getEventsToSend(limit)
		if err != nil {
			return err
		}

		if len(events) > 0 {
			l.logger.Info(fmt.Sprintf("Processing %d events", len(events)))
		}

		for _, e := range events {
			if err := l.publishEvent(e, callback); err != nil {
				return err
			}

			if err := l.setEventAsSent(e); err != nil {
				return err
			}
		}

		l.logger.Info("Polling for new unsent events")
		time.Sleep(l.waitSeconds)
	}
}

func (l *Listener) connectToDatabase(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	l.db = db

	l.logger.Debug("Connected to database", "dsn", dsn)
	return nil
}

func (l *Listener) transaction(callback func(tx *sql.Tx) error) error {
	l.logger.Debug("Opening transaction")

	tx, err := l.db.Begin()
	if err != nil {
		return err
	}

	if err := callback(tx); err != nil {
		if rollErr := tx.Rollback(); rollErr != nil {
			return errors.Join(rollErr)
		}

		return err
	}

	return tx.Commit()
}

func (l *Listener) setupDatabaseSchema(config Config) error {
	return l.transaction(func(tx *sql.Tx) error {
		if err := l.setupEventsTable(tx); err != nil {
			return err
		}

		l.logger.Debug("Schema and trigger setup complete")

		for _, table := range config.Tables {
			if err := l.setupTableTrigger(tx, table); err != nil {
				return err
			}

			l.logger.Debug("Table setup complete", "table", table)
		}

		return nil
	})
}

func (l *Listener) setupTableToChannelRelation(channels map[string]event.Channel) {
	l.tableToChannelRelation = make(map[string][]event.Channel, len(channels))

	for _, channel := range channels {
		l.tableToChannelRelation[channel.From] = append(
			l.tableToChannelRelation[channel.From],
			channel,
		)
	}
}

func (l *Listener) setupEventsTable(tx *sql.Tx) error {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	if _, err := tx.ExecContext(ctx, setupFromToEventTableQuery); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, setupFromToProcessEventFunctionQuery); err != nil {
		return err
	}

	return nil
}

func (l *Listener) setupTableTrigger(tx *sql.Tx, table string) error {
	triggerName := fmt.Sprintf("from_to_%s_process_event_trigger", table)
	query := fmt.Sprintf(setupTableTriggerPartialQuery, triggerName, table)

	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	if _, err := tx.ExecContext(ctx, query); err != nil {
		return err
	}

	return nil
}

func (l *Listener) getEventsToSend(limit uint64) ([]event.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	rows, err := l.db.QueryContext(ctx, getEventsToSendQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		var e event.Event
		var data []byte
		if err := rows.Scan(&e.ID, &e.Op, &e.Table, &data, &e.Ts, &e.Sent); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &e.Row); err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, nil
}

func (l *Listener) setEventAsSent(e event.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	if _, err := l.db.ExecContext(ctx, setEventAsSentQuery, e.ID); err != nil {
		return err
	}

	l.logger.Debug("Marked event as sent", "event", e)

	return nil
}

func (l *Listener) publishEvent(e event.Event, callback func(event.Event, []event.Channel) error) error {
	l.logger.Debug("Publishing event", "event", e)
	l.logger.Debug("Getting channels to publish")

	channels, ok := l.tableToChannelRelation[e.Table]
	if !ok {
		l.logger.Warn("Table does not have any configured channel, skipping", "id", e.ID, "table", e.Table)
		return nil
	}

	l.logger.Debug("Calling publish event callback", "event", e, "channels", channels)

	if err := callback(e, channels); err != nil {
		return fmt.Errorf("Failed to publish event, got error %s", err.Error())
	}

	return nil
}
