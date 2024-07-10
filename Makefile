.PHONY: test test-coverage

test:
	go test ./... -v

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html