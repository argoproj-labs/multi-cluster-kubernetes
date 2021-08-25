install:
	go run ./cmd/mck cluster rm garbage -n default
	go run ./cmd/mck cluster add default -n default
start: install
	go run ./cmd/mck server -n default
test: install
	go test ./...
lint:
	golangci-lint run --fix