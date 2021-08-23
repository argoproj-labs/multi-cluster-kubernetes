install:
	kubectl cluster-info
	kubectl -n default create secret generic clusters --dry-run=client -o yaml | kubectl apply -f -
	go run ./cmd -n default cluster add default docker-desktop
	go run ./cmd -n default cluster add docker-desktop docker-desktop
start:
	go run ./cmd -n default server
test:
	go test ./...