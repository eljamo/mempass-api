ARG TARGETARCH

FROM golang:1.26 AS build

WORKDIR /workspace
COPY . .

ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=${TARGETARCH}

RUN go mod download && go mod verify
RUN go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot-${TARGETARCH}

COPY --from=build /workspace/server /usr/bin/server
CMD ["/usr/bin/server"]
