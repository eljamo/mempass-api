FROM golang:1.21 AS build

WORKDIR /workspace
COPY . .
RUN go mod download
RUN GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:latest

COPY --from=build /workspace/server /usr/bin/server
CMD ["/usr/bin/server"]
