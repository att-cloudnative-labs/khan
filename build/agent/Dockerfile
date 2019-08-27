# ------------------------------------------------------------------------
#  Stage 1 of 2 container, for building go binary
# ------------------------------------------------------------------------
FROM golang:1.12.1-stretch as builder

COPY . /agent

WORKDIR /agent

RUN go mod download && \
    GIT_TERMINAL_PROMPT=1 \
    GOARCH=amd64 \
    GOOS=linux \
    CGO_ENABLED=0 \
    go build -v \
    --installsuffix cgo \
    -o bin/agent \
    cmd/agent/main.go

# ------------------------------------------------------------------------
#  Stage 2 of 2 container, for running application
# ------------------------------------------------------------------------
FROM alpine:3.8
COPY --from=builder /agent/bin/agent /usr/bin/agent

ENTRYPOINT ["/usr/bin/agent"]
CMD ["--logtostderr"]