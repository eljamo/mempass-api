# mempass-api

`mempass-api` is a example RPC service built with [Connect][connect] and [libpass][libpass]. Its API is defined by a [Protocol Buffer schema][schema], and the service
supports the [gRPC][grpc-protocol], [gRPC-Web][grpcweb-protocol], and [Connect protocols][connect-protocol].

## Run

```
go run ./cmd/server
```

## Run on Docker

```
docker compose up
```

## Call using `curl`

```
curl -i \
    --header "Content-Type: application/json" \
     --data "{}" \
    http://127.0.0.1:4321/mempass.v1.PasswordService/GeneratePasswords
```

```bash
curl -i \
    --header "Content-Type: application/json" \
    --data '{"preset": "XKCD", "word_list": "POKEMON"}' \
    http://127.0.0.1:4321/mempass.v1.PasswordService/GeneratePasswords
```

```bash
curl -i \
    --header "Content-Type: application/json" \
    --data '{
        "preset": "XKCD",
        "word_list": "MIDDLE_EARTH",
        "case_transform": "SENTENCE",
        "num_passwords": 10
    }' \
    http://127.0.0.1:4321/mempass.v1.PasswordService/GeneratePasswords
```

## Call using `grpcurl`

```bash
grpcurl \
    -protoset <(buf build -o -) -plaintext \
    -d '{}' \
    127.0.0.1:4321 mempass.v1.PasswordService/GeneratePasswords
```

```bash
grpcurl \
    -protoset <(buf build -o -) -plaintext \
    -d '{"preset": "XKCD", "word_list": "POKEMON"}' \
    127.0.0.1:4321 mempass.v1.PasswordService/GeneratePasswords
```

```bash
grpcurl \
    -protoset <(buf build -o -) -plaintext \
    -d '{
        "preset": "XKCD",
        "word_list": "MIDDLE_EARTH",
        "case_transform": "SENTENCE", 
        "num_passwords": 10
    }' \
    127.0.0.1:4321 mempass.v1.PasswordService/GeneratePasswords
```

[connect]: https://github.com/connectrpc/connect-go
[connect-protocol]: https://connectrpc.com/docs/protocol
[grpc-protocol]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
[grpcweb-protocol]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md
[libpass]: https://github.com/eljamo/libpass
[schema]: https://github.com/eljamo/mempass-api/blob/main/proto/mempass/v1/mempass.proto