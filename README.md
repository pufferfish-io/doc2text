# doc2text

## Run

```
export $(cat .env | xargs) && go run ./cmd/doc2text
```

## Recreate gRPC

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/presentation/proto/ocr/v1/ocr.proto
```
