package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gustapinto/from-to/internal/event"
	"github.com/lib/pq"
)

type Listener struct {
	dsn                    string
	waitSeconds            time.Duration
	timeout                time.Duration
	db                     *sql.DB
	logger                 *slog.Logger
	tableToChannelRelation map[string][]event.Channel
}

func NewListener(config Config, channels map[string]event.Channel) (*Listener, error) {
	waitSeconds := time.Duration(config.PollSeconds) * time.Second
	if waitSeconds == 0 {
		waitSeconds = time.Duration(30) * time.Second
	}

	listener := &Listener{
		dsn:         config.DSN,
		waitSeconds: waitSeconds,
		timeout:     1 * time.Minute,
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

func (c *Listener) Listen(callback func(event.Event) error) error {
	listener := pq.NewListener(c.dsn, 1*time.Second, 10*time.Second, nil)
	defer listener.Close()

	if err := listener.Listen("from_to_event_notifications"); err != nil {
		return err
	}

	for {
		select {
		case n := <-listener.Notify:
			if err := c.handleNotification(n, callback); err != nil {
				c.logger.Error(err.Error())
			}

		case <-time.After(c.waitSeconds):
			c.logger.Debug("Waiting for new rows", "waitSeconds", c.waitSeconds)
		}
	}
}

func (c *Listener) connectToDatabase(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	c.db = db

	c.logger.Debug("Connected to database", "dsn", dsn)
	return nil
}

func (c *Listener) setupDatabaseSchema(config Config) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	if err := c.setupEventsTable(tx); err != nil {
		tx.Rollback()
		return err
	}

	c.logger.Debug("Schema and trigger setup complete")

	for _, table := range config.Tables {
		if err := c.setupTable(tx, table); err != nil {
			tx.Rollback()
			return err
		}

		c.logger.Debug("Table setup complete", "table", table)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *Listener) setupTableToChannelRelation(channels map[string]event.Channel) {
	c.tableToChannelRelation = make(map[string][]event.Channel, len(channels))

	for _, channel := range channels {
		c.tableToChannelRelation[channel.From] = append(
			c.tableToChannelRelation[channel.From],
			channel,
		)
	}
}

func (c *Listener) setupEventsTable(tx *sql.Tx) error {
	const query = `
	CREATE TABLE IF NOT EXISTS "from_to_event" (
		"id" BIGSERIAL PRIMARY KEY,
		"op" CHAR(1) NOT NULL,
		"table" VARCHAR(255) NOT NULL,
		"row" JSONB NOT NULL,
		"ts" BIGINT NOT NULL
	);

	CREATE OR REPLACE FUNCTION "from_to_process_event"()
	RETURNS TRIGGER
	AS $$
	BEGIN
		IF (TG_OP = 'DELETE') THEN
			INSERT INTO "from_to_event" (
				"op",
				"table",
				"row",
				"ts"
			)
			SELECT
				'D',
				TG_TABLE_NAME,
				row_to_json(OLD.*),
				(extract(epoch from now()));
		ELSIF (TG_OP = 'UPDATE') THEN
			INSERT INTO "from_to_event" (
				"op",
				"table",
				"row",
				"ts"
			)
			SELECT
				'U',
				TG_TABLE_NAME,
				row_to_json(NEW.*),
				(extract(epoch from now()));
		ELSIF (TG_OP = 'INSERT') THEN
			INSERT INTO "from_to_event" (
				"op",
				"table",
				"row",
				"ts"
			)
			SELECT
				'I',
				TG_TABLE_NAME,
				row_to_json(NEW.*),
				(extract(epoch from now()));
		END IF;
		RETURN NULL;
	END
	$$ LANGUAGE PLPGSQL;

	CREATE OR REPLACE FUNCTION "from_to_notify_event"()
	RETURNS TRIGGER as $$
	BEGIN
		PERFORM PG_NOTIFY('from_to_event_notifications', NEW.id::text);
		RETURN NULL;
	END
	$$ LANGUAGE PLPGSQL;

	CREATE OR REPLACE TRIGGER "from_to_event_notidy_event_trigger"
	AFTER INSERT ON "from_to_event"
	FOR EACH ROW EXECUTE FUNCTION from_to_notify_event();
	`

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Listener) setupTable(tx *sql.Tx, table string) error {
	query := `
	CREATE OR REPLACE TRIGGER %s
	AFTER INSERT OR UPDATE OR DELETE ON %s
	FOR EACH ROW EXECUTE FUNCTION from_to_process_event()
	`

	triggerName := fmt.Sprintf("from_to_%s_process_event_trigger", table)
	query = fmt.Sprintf(query, triggerName, table)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Listener) handleNotification(n *pq.Notification, callback func(event.Event) error) error {
	if n == nil {
		return fmt.Errorf("Failed to process nil row")
	}

	var id int64
	if err := json.Unmarshal([]byte(n.Extra), &id); err != nil {
		return fmt.Errorf("Failed to process received row, got error %s", err.Error())
	}

	c.logger.Info("Processing received row", "id", id)

	event, err := c.getEvent(id)
	if err != nil {
		return fmt.Errorf("Failed to process received row, got error %s", err.Error())
	}

	c.logger.Debug("Row processed into event", "event", event)

	if err := callback(event); err != nil {
		return fmt.Errorf("Failed to publish event, got error %s", err.Error())
	}

	return nil
}

func (c *Listener) getEvent(id int64) (e event.Event, err error) {
	const query = `
	SELECT
		fte.id,
		fte.op,
		fte.table,
		fte.row,
		fte.ts
	FROM
		from_to_event fte
	WHERE
		fte.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	row := c.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		return event.Event{}, row.Err()
	}

	var data []byte
	if err := row.Scan(&e.ID, &e.Op, &e.Table, &data, &e.Ts); err != nil {
		return event.Event{}, err
	}

	if err := json.Unmarshal(data, &e.Row); err != nil {
		return event.Event{}, err
	}

	if channels, exists := c.tableToChannelRelation[e.Table]; exists {
		copy(e.Channels, channels)
	}

	return e, nil
}
