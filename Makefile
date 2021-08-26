install:
	go run ./cmd/mck config rm garbage -n default
	go run ./cmd/mck config add default -n default
start: install
	go run ./cmd/mck server -n default
test: install
	go test ./...
lint:
	golangci-lint run --fix