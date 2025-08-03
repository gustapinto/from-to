package postgres

const (
	setupFromToEventTableQuery = `
	CREATE TABLE IF NOT EXISTS "from_to_event" (
		"id" BIGSERIAL PRIMARY KEY,
		"op" CHAR(1) NOT NULL,
		"table" VARCHAR(255) NOT NULL,
		"row" JSONB NOT NULL,
		"ts" BIGINT NOT NULL,
		"sent" BOOLEAN NOT NULL DEFAULT FALSE
	);
	`

	setupFromToProcessEventFunctionQuery = `
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
	`

	setupTableTriggerPartialQuery = `
	CREATE OR REPLACE TRIGGER %s
	AFTER INSERT OR UPDATE OR DELETE ON %s
	FOR EACH ROW EXECUTE FUNCTION from_to_process_event()
	`

	getEventsToSendQuery = `
	SELECT
		fte.id,
		fte.op,
		fte.table,
		fte.row,
		fte.ts,
		fte.sent
	FROM
		from_to_event fte
	WHERE
		fte.sent = FALSE
	ORDER BY
		fte.ts ASC
	LIMIT
		$1::BIGINT
	`

	setEventAsSentQuery = `
	UPDATE
		from_to_event
	SET
		sent = TRUE
	WHERE
		id = $1
	`
)
