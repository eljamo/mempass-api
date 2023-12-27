FROM golang:1.21-alpine AS builder

WORKDIR /workspace

RUN apk add --update --no-cache git && rm -rf /var/cache/apk/*
COPY go.mod go.sum /workspace/
RUN go mod download
COPY cmd /workspace/cmd
COPY internal /workspace/internal
RUN go build -o mempass-server ./cmd/server

FROM alpine:3.19
RUN apk add --update --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*
COPY --from=builder /workspace/mempass-server /usr/local/bin/mempass-server
CMD [ "/usr/local/bin/mempass-server", "--server-stream-delay=500ms" ]
