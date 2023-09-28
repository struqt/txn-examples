#
# https://atlasgo.io/atlas-schema/hcl
#

schema "example" {
  comment = "Schema of the example"
}

table "authors" {
  schema  = schema.example
  comment = "Author Table"
  column "id" {
    null = false
    type = bigserial
  }
  primary_key {
    columns = [column.id]
  }
  column "name" {
    comment = "Full name of the author"
    null    = false
    type    = varchar
  }
  column "bio" {
    null = true
    type = text
  }
  column "created_at" {
    null    = true
    type    = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  index "authors_created_at_k" {
    columns = [column.created_at]
  }
}
