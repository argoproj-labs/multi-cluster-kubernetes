install:
	go run ./cmd/mck cluster rm garbage
	go run ./cmd/mck cluster add
start:
	go run ./cmd/mck server
test: install
	go test ./...