# Examples

[![Go Build](https://github.com/struqt/txn-examples/actions/workflows/go.yml/badge.svg)](https://github.com/struqt/txn-examples/actions/workflows/go.yml)

## Build release version

```shell
bash build.sh
```

## Setup development environment

1. Start database servers
2. Finish the `DDL` works
3. Set two environment variables: `DB_HOST`, `DB_PASSWORD`

## Run go modules

```shell
go run examples/sqlc/pgx
```

```shell
go run examples/sqlc/pg
```

```shell
go run examples/sqlc/mysql
```
