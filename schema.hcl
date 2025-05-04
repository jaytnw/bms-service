CREATE TABLE "statuses" ("id" bigserial,"drom_id" varchar(100) NOT NULL,"washing_id" varchar(100) NOT NULL,"status" varchar(50) NOT NULL,"created_at" timestamptz,"updated_at" timestamptz,PRIMARY KEY ("id"));
CREATE INDEX IF NOT EXISTS "idx_statuses_washing_id" ON "statuses" ("washing_id");
CREATE INDEX IF NOT EXISTS "idx_statuses_id" ON "statuses" ("id");

