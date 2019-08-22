# ------------------------------------------------------------------------
#  The first stage container, for setting up and downloading dependencies
# ------------------------------------------------------------------------
FROM golang:1.12.1-stretch as builder1

#COPY .gitconfig /root/.gitconfig
COPY . /app

# Add the keys
ARG bitbucket_id
ENV bitbucket_id=$bitbucket_id
ARG bitbucket_token
ENV bitbucket_token=$bitbucket_token

# set up environment
ARG env_preset
ENV PRESET=$env_preset

WORKDIR /app/cmd/webapp

RUN git config --global url."https://${bitbucket_id}:${bitbucket_token}@egbitbucket.dtvops.net/".insteadOf "https://egbitbucket.dtvops.net/" &&\
    go mod download
    
# ------------------------------------------------------------------------
#  The second stage container, for building the Go Binary
# ------------------------------------------------------------------------
FROM builder1 AS builder2
COPY . /app
WORKDIR /app/cmd/webapp

RUN GOATT_VERSION=$(grep -e goatt ../../go.mod | cut -d' ' -f2) &&\
    GIT_TERMINAL_PROMPT=1 \
    GOARCH=amd64 \
    GOOS=linux \
    CGO_ENABLED=0 \
    go build -v \
    --installsuffix cgo \
    --ldflags "-X main.goattVersion=$GOATT_VERSION -s" \
    -o controller

# ------------------------------------------------------------------------
#  The third stage container, for running the application
# ------------------------------------------------------------------------
FROM alpine:3.8
COPY --from=builder2 /app/cmd/webapp/controller /app/controller
COPY --from=builder2 /app/settings.yaml /app/settings.yaml

ENV SERVER_SSL_ENABLED=true
ENV ENABLE_PROFILING=false

RUN apk update && apk add --no-cache ca-certificates

# USER app

WORKDIR /app

ENTRYPOINT ["/app/controller"]