VERSION ?= $(shell git describe --exact-match --tags $$(git log -n1 --pretty='%h') 2>/dev/null || echo "$$(git rev-parse --abbrev-ref HEAD)-$$(git rev-parse --short HEAD)")
CONTAINER_IMAGE := ghcr.io/kakkoyun/kafkaques:$(VERSION)

LDFLAGS="-X main.version=$(VERSION)"

.PHONY: build
build: bin/kafkaques

bin/kafkaques: deps main.go
	mkdir -p bin
	go build -a -ldflags=$(LDFLAGS) -o $@ .

.PHONY: clean
clean:
	rm -rf bin

.PHONY: deps
deps: go.mod
	go mod tidy

.PHONY: format
	go fmt `go list ./...`

.PHONY: test
test:
	 go test -v `go list ./...`

.PHONY: container
container: kafkaques
	docker build -t $(CONTAINER_IMAGE) --build-arg VERSION=$(VERSION) .

.PHONY: push-container
push-container:
	docker push $(CONTAINER_IMAGE)
