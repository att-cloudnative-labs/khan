VERSION=v1.1.0

.PHONY: registry agent test

all: registry agent

registry: test
	docker build -f build/registry/Dockerfile -t $(DOCKER_REGISTRY)/khan-system/khan-registry:$(VERSION) .

agent: test
	docker build -f build/agent/Dockerfile -t $(DOCKER_REGISTRY)/khan-system/khan-agent:$(VERSION) .

test:
	go test -v ./...