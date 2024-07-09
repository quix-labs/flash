module github.com/quix-labs/flash/drivers/trigger

go 1.21.6

replace github.com/quix-labs/flash => ../../

require (
	github.com/lib/pq v1.10.9
	github.com/quix-labs/flash v0.0.0-00010101000000-000000000000
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
)
