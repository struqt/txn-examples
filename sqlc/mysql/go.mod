module examples/sqlc/mysql

go 1.20

require github.com/struqt/txn v0.0.0

replace github.com/struqt/txn => ../../txn

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/struqt/x v0.3.1
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
