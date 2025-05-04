-- Drop index "idx_statuses_washer_id" from table: "statuses"
DROP INDEX "public"."idx_statuses_washer_id";
-- Create index "idx_washer_created" to table: "statuses"
CREATE INDEX "idx_washer_created" ON "public"."statuses" ("washer_id", "created_at");
