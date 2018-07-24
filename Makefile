default: clean build

.PHONY: build
build: clean loadtest-collector

loadtest-collector:
	CGO_ENABLED=0 GOOS=linux go build

.PHONY: clean
clean:
	-rm loadtest-collector

.PHONY: docker
docker:
	docker build -t jspc/loadtest-collector .
	docker push jspc/loadtest-collector
