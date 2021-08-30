install:
	kubectl -n default delete secret -l multi-cluster.argoproj.io/kubeconfig
	# test we can run without 404 error
	go run ./cmd/mck config get -n default default
	go run ./cmd/mck config add -n default default
	# test we do not get 409 error
	go run ./cmd/mck config add -n default default
	go run ./cmd/mck config get -n default default
start: install
	go run ./cmd/mck server -n default
test: install
	go test ./...
lint:
	golangci-lint run --fix