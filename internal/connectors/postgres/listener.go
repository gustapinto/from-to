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

type Postgres struct {
	dsn           string
	waitSeconds   time.Duration
	db            *sql.DB
	logger        *slog.Logger
	tableRelation map[string]SetupParamsTable
}

func NewListener(params SetupParams) (c *Postgres, err error) {
	c = &Postgres{
		dsn:           params.DSN,
		waitSeconds:   time.Duration(params.PollSeconds) * time.Second,
		logger:        slog.With("listener", "Postgres"),
		tableRelation: make(map[string]SetupParamsTable),
	}

	c.db, err = sql.Open("postgres", params.DSN)
	if err != nil {
		return nil, err
	}

	if err := c.db.Ping(); err != nil {
		return nil, err
	}

	c.logger.Debug("Connected to database", "dsn", params.DSN)

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	if err := c.setupEventsTable(tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	c.logger.Debug("Schema and trigger setup complete")

	for _, table := range params.Tables {
		if c.tableRelation == nil {
			c.tableRelation = make(map[string]SetupParamsTable)
		}

		if err := c.setupTable(tx, table); err != nil {
			tx.Rollback()
			return nil, err
		}

		c.logger.Debug("Table setup complete", "table", table.Name)
		c.tableRelation[table.Name] = table
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	c.logger.Info("Connector setup completed")

	return c, nil
}

// func (c *Postgres) Listen(out output.Connector) error {
func (c *Postgres) Listen(callback func(event.Event) error) error {
	listener := pq.NewListener(c.dsn, 1*time.Second, 10*time.Second, nil)
	defer listener.Close()

	if err := listener.Listen("from_to_event_notifications"); err != nil {
		return err
	}

	for {
		select {
		case n := <-listener.Notify:
			var id int64
			if err := json.Unmarshal([]byte(n.Extra), &id); err != nil {
				c.logger.Error("Failed to process received row", "error", err)
				continue
			}

			c.logger.Info("Processing received row", "id", id)

			event, err := c.getEvent(id)
			if err != nil {
				c.logger.Error("Failed to process received row", "error", err)
				continue
			}

			c.logger.Debug("Row processed into event", "event", event)

			if err := callback(event); err != nil {
				c.logger.Error("Failed to publish event", "error", err)
				continue
			}

			c.logger.Debug("Event published", "event", event)
			break

		case <-time.After(c.waitSeconds):
			c.logger.Debug("Waiting for new rows", "waitSeconds", c.waitSeconds)
			break
		}
	}
}

func (c *Postgres) setupEventsTable(tx *sql.Tx) error {
	query := `
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
			INSERT INTO "from_to_event" ("op", "table", "row", "ts") SELECT 'D', TG_TABLE_NAME, row_to_json(OLD.*), (extract(epoch from now()));
		ELSIF (TG_OP = 'UPDATE') THEN
			INSERT INTO "from_to_event" ("op", "table", "row", "ts") SELECT 'U', TG_TABLE_NAME, row_to_json(NEW.*), (extract(epoch from now()));
		ELSIF (TG_OP = 'INSERT') THEN
			INSERT INTO "from_to_event" ("op", "table", "row", "ts") SELECT 'I', TG_TABLE_NAME, row_to_json(NEW.*), (extract(epoch from now()));
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

	_, err := tx.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Postgres) setupTable(tx *sql.Tx, table SetupParamsTable) error {
	query := `
	CREATE OR REPLACE TRIGGER %s
	AFTER INSERT OR UPDATE OR DELETE ON %s
	FOR EACH ROW EXECUTE FUNCTION from_to_process_event()
	`

	triggerName := fmt.Sprintf("from_to_%s_process_event_trigger", table.Name)
	query = fmt.Sprintf(query, triggerName, table.Name)

	_, err := tx.ExecContext(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Postgres) getEvent(id int64) (e event.Event, err error) {
	query := `
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

	row := c.db.QueryRowContext(context.Background(), query, id)
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

	if table, exists := c.tableRelation[e.Table]; exists {
		e.Metadata = event.EventMetadata{
			Key:      table.KeyColumn,
			KeyValue: fmt.Sprint(e.Row[table.KeyColumn]),
			Topic:    table.Topic,
		}
	}

	return e, nil
}
