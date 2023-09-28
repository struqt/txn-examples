#
# https://atlasgo.io/atlas-schema/hcl
#

schema "example" {
  comment = "Example Schema"
  charset = "utf8mb4"
  collate = "utf8mb4_general_ci"
}

table "authors" {
  schema  = schema.example
  comment = "Author Table"
  column "id" {
    null           = false
    type           = bigint
    auto_increment = true
  }
  primary_key {
    columns = [column.id]
  }
  column "name" {
    comment = "Full name of the author"
    null    = false
    type    = varchar(500)
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
