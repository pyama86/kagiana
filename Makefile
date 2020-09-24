VERSION  := $(shell git tag | tail -n1 | sed 's/v//g')
TEST ?= $(shell go list ./... | grep -v -e vendor)
test:
	go test -v $(TEST)

docker:
	docker build -t pyama/kagiana:$(VERSION) .
	docker push pyama/kagiana:$(VERSION)
