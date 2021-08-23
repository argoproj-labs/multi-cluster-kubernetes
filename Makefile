install:
	go run ./cmd/mck -n default cluster add default docker-desktop
start:
	go run ./cmd/mck -n default server
test: install
	go test ./...