# Использование и запуск

Этот документ описывает, как пользоваться doc2text и как его запускать локально и в Docker.

Что делает сервис
- Принимает запрос по gRPC с `objectkey`.
- Скачивает файл из S3/MinIO по ключу.
- Конвертирует файл в Base64 и отправляет в Yandex OCR.
- Возвращает распознанный текст.

Требования
- Go (версия согласно `go.mod`, сейчас `1.24`+)
- Docker (опционально)
- grpcurl (опционально для ручных вызовов)

Переменные окружения (основные)
- gRPC: `G_RPC_SERVER_DOC2TEXT_ADDR` (по умолчанию `:8080`)
- HTTP‑health: `HTTP_SERVER_DOC2TEXT_ADDR` (по умолчанию `:8090`)
- S3/MinIO: `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`, `S3_USE_SSL`
- Yandex OCR: `YC_API_KEY`, `YC_FOLDER_ID`, `YC_ENDPOINT` (по умолчанию batchAnalyze), `YC_DEFAULT_MODEL`, `YC_LANGUAGES`, `YC_MIN_CONFIDENCE`, `YC_HTTP_TIMEOUT`
- OIDC (необязательно): `OIDC_DOC2TEXT_ISSUER`, `OIDC_DOC2TEXT_JWKS_URL`, `OIDC_DOC2TEXT_AUDIENCE`, `OIDC_DOC2TEXT_EXPECTED_AZP`

Пример `.env`
```
G_RPC_SERVER_DOC2TEXT_ADDR=:50051
HTTP_SERVER_DOC2TEXT_ADDR=:8090

S3_ENDPOINT=minio.example.local:9000
S3_ACCESS_KEY=ACCESS
S3_SECRET_KEY=SECRET
S3_BUCKET=files
S3_USE_SSL=false

YC_API_KEY=***
YC_FOLDER_ID=***
YC_ENDPOINT=https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze
YC_DEFAULT_MODEL=page
YC_LANGUAGES=ru,en
YC_MIN_CONFIDENCE=0.6
YC_HTTP_TIMEOUT=15s

# опционально (включает проверку JWT для gRPC)
OIDC_DOC2TEXT_ISSUER=https://auth.example.com/realms/demo
OIDC_DOC2TEXT_JWKS_URL=https://auth.example.com/realms/demo/protocol/openid-connect/certs
OIDC_DOC2TEXT_AUDIENCE=doc2text
OIDC_DOC2TEXT_EXPECTED_AZP=my-client
```

Запуск локально (Go)
```
set -a && source .env && set +a
go run ./cmd/doc2text
```

Проверка здоровья (HTTP)
```
curl http://localhost:8090/healthz
```

Вызов gRPC через grpcurl
```
grpcurl -plaintext -d '{"objectkey":"folder/file.pdf"}' \
  localhost:50051 ocr.v1.OcrService/Process
```
Если включён OIDC, добавьте заголовок авторизации:
```
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{"objectkey":"folder/file.pdf"}' \
  localhost:50051 ocr.v1.OcrService/Process
```

Docker: сборка и запуск
```
docker build -t doc2text:dev .
docker run --rm -p 50051:50051 -p 8090:8090 \
  --env-file ./.env doc2text:dev
```

Docker: образ из GHCR
- Теги: версии (`vX.Y.Z`) и `latest`.
```
docker pull ghcr.io/pufferfish-io/doc2text:latest
docker run --rm -p 50051:50051 -p 8090:8090 \
  --env-file ./.env ghcr.io/pufferfish-io/doc2text:latest
```

Пересборка protobuf (при изменении .proto)
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  internal/presentation/proto/ocr/v1/ocr.proto
```
