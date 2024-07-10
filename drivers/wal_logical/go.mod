module github.com/quix-labs/flash/drivers/wal_logical

go 1.21.6

replace github.com/quix-labs/flash => ../../

require (
	github.com/jackc/pglogrepl v0.0.0-20240307033717-828fbfe908e9
	github.com/jackc/pgx/v5 v5.6.0
	github.com/quix-labs/flash v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
