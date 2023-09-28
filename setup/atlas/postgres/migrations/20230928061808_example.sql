-- Modify "authors" table
ALTER TABLE "example"."authors" ADD COLUMN "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP;
-- Create index "authors_created_at_k" to table: "authors"
CREATE INDEX "authors_created_at_k" ON "example"."authors" ("created_at");
