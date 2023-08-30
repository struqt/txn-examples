module examples/sqlc/mysql

go 1.20

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/struqt/logging v0.0.0-20230830051957-37f9d79d2d35
	github.com/struqt/txn v0.0.0-20230830051924-1c346c53d0d1
)

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/zerologr v1.2.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.30.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace (
	github.com/struqt/logging => ../../logging
	github.com/struqt/txn => ../../txn
)
