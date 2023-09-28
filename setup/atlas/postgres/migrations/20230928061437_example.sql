-- Add new schema named "example"
CREATE SCHEMA "example";
-- Set comment to schema: "example"
COMMENT ON SCHEMA "example" IS 'Schema of the example';
-- Create "authors" table
CREATE TABLE "example"."authors" ("id" bigserial NOT NULL, "name" character varying NOT NULL, "bio" text NULL, PRIMARY KEY ("id"));
-- Set comment to table: "authors"
COMMENT ON TABLE "example"."authors" IS 'Author Table';
-- Set comment to column: "name" on table: "authors"
COMMENT ON COLUMN "example"."authors" ."name" IS 'Full name of the author';
