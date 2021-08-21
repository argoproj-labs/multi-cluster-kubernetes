build:
	kubectl -n argo create secret generic clusters --dry-run=client -o yaml | kubectl apply -f -
	go run ./cmd cluster add -n argo default docker-desktop
	go run ./cmd server -n argo
	# GODEBUG=http2debug=2 go run ./cmd server -n argo