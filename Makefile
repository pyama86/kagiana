VERSION  := $(shell git tag | tail -n1 | sed 's/v//g')
TEST ?= $(shell go list ./... | grep -v -e vendor)
test:
	go mod tidy
	go test -v $(TEST)

docker:
	docker build --platform linux/amd64 -t pyama/kagiana:$(VERSION) .
	docker push pyama/kagiana:$(VERSION)
	docker tag pyama/kagiana:$(VERSION) pyama/kagiana:latest
	docker push pyama/kagiana:latest

.PHONY: release_major
## release_major: release nke (major)
release_major:
	git semv major --bump

.PHONY: release_minor
## release_minor: release nke (minor)
release_minor:
	git semv minor --bump

.PHONY: release_patch
## release_patch: release nke (patch)
release_patch:
	git semv patch --bump

release_deps:
	which goreleaser > /dev/null || go install github.com/goreleaser/goreleaser@latest

release: release_deps
	goreleaser --rm-dist --skip-validate 
