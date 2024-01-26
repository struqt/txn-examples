# Examples

[![Go Build](https://github.com/struqt/txn-examples/actions/workflows/go.yml/badge.svg)](https://github.com/struqt/txn-examples/actions/workflows/go.yml)

## Build release version

```shell
bash build.sh
```

## Setup development environment

1. Start database servers
2. Finish the `DDL` works

## Run go modules

Set environment variables: `DB_ADDR_TCP`, `DB_PASSWORD`

```shell
go run examples/sqlc/pgx
```

```shell
go run examples/sqlc/pg
```

```shell
go run examples/sqlc/mysql
```

```shell
go run examples/sqlc/sqlite
```

Set environment variable: `DB_ADDR_PATH`

```shell
go run examples/mongo
```
