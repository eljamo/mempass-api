ARG GO_ARCH=arm64

FROM golang:1.21 AS build

WORKDIR /workspace
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${GO_ARCH}

RUN go mod download && go mod verify

RUN go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot-${GO_ARCH}

COPY --from=build /workspace/server /usr/bin/server
CMD ["/usr/bin/server"]
