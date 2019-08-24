VERSION=v1.0.0-BETA-1

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

all: registry agent

registry:
	go build -o bin/registry cmd/registry/main.go

agent:
	go build -o bin/agent cmd/agent/main.go
