install:
	go run ./cmd -n default cluster add default docker-desktop
start:
	go run ./cmd -n default server
test: install
	go test ./...