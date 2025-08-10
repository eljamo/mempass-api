ARG GOARCH

FROM golang:1.24 AS build

WORKDIR /workspace
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${GOARCH}

RUN go mod tidy && go mod download && go mod verify && go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot-${GOARCH}

COPY --from=build /workspace/server /usr/bin/server
CMD ["/usr/bin/server"]
