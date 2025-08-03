BEGIN;

CREATE TABLE IF NOT EXISTS "sales" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP,
    "description" VARCHAR(255) NOT NULL,
    "total_value" DECIMAL(10, 2) NOT NULL
);

COMMIT;
