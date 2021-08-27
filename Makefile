install:
	kubectl delete secret kubeconfig -n default --ignore-not-found
	# test we can run without 404 error
	go run ./cmd/mck config get -n default
	go run ./cmd/mck config add -n default
	# test we do not get 409 error
	go run ./cmd/mck config add -n default
	go run ./cmd/mck config get -n default
start: install
	go run ./cmd/mck server -n default
test: install
	go test ./...
lint:
	golangci-lint run --fix