install:
	kubectl cluster-info
	kubectl create secret generic clusters --dry-run=client -o yaml | kubectl apply -f -
	go run ./cmd cluster add default docker-desktop
start:
	go run ./cmd server
test:
	go test ./...