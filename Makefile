.PHONY: test test-coverage

test:
	go test -v ./... ./drivers/trigger/ ./drivers/wal_logical/

test-coverage:
	go test -coverprofile=coverage.out ./... ./drivers/trigger/ ./drivers/wal_logical/
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html
