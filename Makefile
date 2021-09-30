# Help
.PHONY: default
default:
	@echo "Please specify a build target. The choices are:"
	@echo "    test:            Run unit tests"
	@echo "    lint:            Run linting checks"

.PHONY: test
test:
	@echo "============= Running unit tests ============="
	./hack/makecmd test

.PHONY: lint
lint:
	@echo "============= Running linting checks ============="
	./hack/makecmd lint
