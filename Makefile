.PHONY: test lint bench cover clean

## Run tests with race detector
test:
	go test -race -count=1 -v ./...

## Run golangci-lint
lint:
	golangci-lint run

## Run benchmarks
bench:
	go test -bench=. -benchmem -run=^$$ ./...

## Generate coverage report
cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

## Remove generated files
clean:
	rm -f coverage.out coverage.html
