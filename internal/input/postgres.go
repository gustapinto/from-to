package input

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gustapinto/from-to/internal"

	"github.com/lib/pq"
)

type PostgresSetupParamsTable struct {
	Name      string
	KeyColumn string
	Topic     string
}

type PostgresSetupParams struct {
	DSN         string
	PollSeconds int64
	Tables      []PostgresSetupParamsTable
}

type PostgresInputConnector struct {
	dsn           string
	waitSeconds   time.Duration
	db            *sql.DB
	logger        *slog.Logger
	tableRelation map[string]PostgresSetupParamsTable
}

func (c *PostgresInputConnector) Setup(config any) error {
	params, ok := config.(*PostgresSetupParams)
	if !ok {
		return errors.New("Invalid config type passed to PostgresInputConnector.Setup(...), expected *PostgresSetupParamsTable")
	}

	db, err := sql.Open("postgres", params.DSN)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	c.dsn = params.DSN
	c.waitSeconds = time.Duration(params.PollSeconds) * time.Second
	c.db = db
	c.logger = slog.With("connector", "PostgresInputConnector")

	c.logger.Debug("Connected to database", "dsn", params.DSN)

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	if err := c.setupEventsTable(tx); err != nil {
		tx.Rollback()
		return err
	}

	c.logger.Debug("Schema and trigger setup complete")

	for _, table := range params.Tables {
		if c.tableRelation == nil {
			c.tableRelation = make(map[string]PostgresSetupParamsTable)
		}

		if err := c.setupTable(tx, table); err != nil {
			tx.Rollback()
			return err
		}

		c.logger.Debug("Table setup complete", "table", table.Name)
		c.tableRelation[table.Name] = table
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	c.logger.Info("Connector setup completed")

	return nil
}

func (c *PostgresInputConnector) Listen(out internal.OutputConnector) error {
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

			if err := out.Publish(event); err != nil {
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

func (c *PostgresInputConnector) setupEventsTable(tx *sql.Tx) error {
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

func (c *PostgresInputConnector) setupTable(tx *sql.Tx, table PostgresSetupParamsTable) error {
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

func (c *PostgresInputConnector) getEvent(id int64) (event internal.Event, err error) {
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
		return internal.Event{}, row.Err()
	}

	var data []byte
	if err := row.Scan(&event.ID, &event.Op, &event.Table, &data, &event.Ts); err != nil {
		return internal.Event{}, err
	}

	if err := json.Unmarshal(data, &event.Row); err != nil {
		return internal.Event{}, err
	}

	if table, exists := c.tableRelation[event.Table]; exists {
		event.Metadata = internal.EventMetadata{
			Key:      table.KeyColumn,
			KeyValue: fmt.Sprint(event.Row[table.KeyColumn]),
			Topic:    table.Topic,
		}
	}

	return event, nil
}
