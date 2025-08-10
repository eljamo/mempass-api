ARG GOARCH

FROM golang:1.24 AS BUILD_IMAGE

WORKDIR /workspace
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${GOARCH}

RUN go mod download && go mod verify
RUN go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot-${GOARCH}

COPY --from=BUILD_IMAGE /workspace/server /usr/bin/server
CMD ["/usr/bin/server"]
