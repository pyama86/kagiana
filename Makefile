
TEST ?= $(shell go list ./... | grep -v -e vendor)
test:
	go test -v $(TEST)

