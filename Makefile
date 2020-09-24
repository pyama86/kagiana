VERSION  := $(shell git tag | tail -n1 | sed 's/v//g')
TEST ?= $(shell go list ./... | grep -v -e vendor)
test:
	go test -v $(TEST)

docker:
	docker build -t pyama86/kagiana:$(VERSION) .
	docker push pyama86/kagiana:$(VERSION)
