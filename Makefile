# Help
.PHONY: default
default:
	@echo "Please specify a build target. The choices are:"
	@echo "    test:            Run unit tests"
	@echo "    build:           Build go code"
	@echo "    lint:            Run linting checks"

.PHONY: test
test:
	@echo "============= Running unit tests ============="
	./hack/makecmd test

.PHONY: build
build:
	@echo "============= Building go code ============="
	./hack/makecmd build

.PHONY: lint
lint:
	@echo "============= Running linting checks ============="
	./hack/makecmd lint
