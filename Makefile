default: clean test build docker

.PHONY: build
build: clean loadtest-collector

loadtest-collector:
	CGO_ENABLED=0 GOOS=linux go build

.PHONY: clean
clean:
	-rm loadtest-collector

.PHONY: test
test: deps
	go test -v -covermode=count -coverprofile="./count.out" ./...

.PHONY: deps
deps:
	go get -u -v ./...

.PHONY: docker
docker:
	docker build -t jspc/loadtest-collector .
	docker push jspc/loadtest-collector
