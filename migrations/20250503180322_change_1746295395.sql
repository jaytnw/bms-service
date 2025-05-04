-- Create "statuses" table
CREATE TABLE "public"."statuses" (
  "id" bigserial NOT NULL,
  "dorm_id" character varying(100) NOT NULL,
  "washer_id" character varying(100) NOT NULL,
  "status" character varying(50) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_statuses_id" to table: "statuses"
CREATE INDEX "idx_statuses_id" ON "public"."statuses" ("id");
-- Create index "idx_statuses_washer_id" to table: "statuses"
CREATE INDEX "idx_statuses_washer_id" ON "public"."statuses" ("washer_id");
