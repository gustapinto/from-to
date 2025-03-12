create table if not exists "sales" (
	id bigserial primary key,
	code uuid unique default gen_random_uuid(),
	salesperson_code uuid not null unique,
	value decimal(10, 2),
	created_at timestamp not null default now()
);

create table if not exists "cdc_events" (
	"id" bigserial primary key,
	"op" cdc_event_op not null,
	"table" varchar(255) not null,
	"row" jsonb not null,
	"ts" timestamp not null default now()
);

create or replace function "process_cdc"()
returns trigger
as $$
begin
	IF (TG_OP = 'DELETE') THEN
        INSERT INTO cdc_events ("op", "table", "row") SELECT 'D', TG_TABLE_NAME, row_to_json(OLD.*);
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO cdc_events ("op", "table", "row") SELECT 'U', TG_TABLE_NAME, row_to_json(NEW.*);
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO cdc_events ("op", "table", "row") SELECT 'I', TG_TABLE_NAME, row_to_json(NEW.*);
    END IF;
    RETURN NULL;
end
$$ language plpgsql;

create or replace function "notify_cdc"() returns trigger as $$
declare
	channel text := tg_argv[0]::text;
begin
	select pg_notify(channel, NEW.id::text);
    RETURN NULL;
end
$$ language plpgsql;

create or replace FUNCTION "setup_cdc_trigger"(table_name varchar)
RETURNS void
AS $$
begin
	execute format(
		'CREATE OR REPLACE TRIGGER "%s" AFTER INSERT OR UPDATE OR DELETE ON "%s" FOR EACH ROW EXECUTE FUNCTION process_cdc()',
		(table_name || '_cdc_trigger'),
		table_name
	);
--    CREATE TRIGGER IF NOT EXISTS (table_name || "_cdc_trigger")
--	AFTER INSERT OR UPDATE OR DELETE ON table_name
--	FOR EACH ROW EXECUTE FUNCTION process_cdc();
end
$$ LANGUAGE plpgsql;

select setup_cdc_trigger('sales');

CREATE or replace TRIGGER "sales_cdc"
AFTER INSERT OR UPDATE OR DELETE ON "sales"
FOR EACH ROW EXECUTE FUNCTION process_cdc();

insert into sales (salesperson_code, value) values (gen_random_uuid(), 15);

create or replace trigger "cdc_events_notifier"
after insert on "cdc_events"
for each row execute function notify_cdc();