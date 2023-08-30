module examples/sqlc/pgx

go 1.20

require (
	github.com/jackc/pgx/v5 v5.4.3
	github.com/struqt/logging v0.0.0-20230830051957-37f9d79d2d35
	github.com/struqt/txn/txn_pgx v0.0.0-20230830051924-1c346c53d0d1
)

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/zerologr v1.2.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.30.0 // indirect
	github.com/struqt/txn v0.0.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace (
	github.com/struqt/logging => ../../logging
	github.com/struqt/txn => ../../txn
	github.com/struqt/txn/txn_pgx => ../../txn/txn_pgx
)
