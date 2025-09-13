# doc2text

**Overview**

- Extracts text from files via OCR and returns it over gRPC.
- Flow: `objectkey` → download from S3 → base64 conversion → call Yandex OCR → return text.

## Architecture (Clean Architecture)

- Goals: separation of concerns, independence from frameworks, testability, clear boundaries, and inward-only dependencies (Dependency Rule).
- Layers in this repo:
  - Domain Contracts: `internal/core/abstraction/*` define pure interfaces (`convert.FileConverter`, `download.Downloader`, `recognize.Recognizer`, `logger.Logger`) and a lightweight CQRS bus.
  - Use Cases: `internal/core/usecase/extracttext` orchestrates downloading, converting to base64, and recognizing text. No external deps imported directly.
  - Infrastructure Adapters: `internal/infrastructure/adapter/*` provide implementations (S3, Yandex OCR, base64 converter, zap logger). They depend on core interfaces, never the other way around.
  - Presentation: `internal/presentation/server/ocr/v1` (gRPC), `internal/presentation/proto/ocr/v1` (proto), `internal/api/router.go` (HTTP health), `internal/presentation/config` (env config).
  - Composition Root: `cmd/doc2text/main.go` wires implementations to interfaces and starts servers.
- CQRS Bus: `internal/core/abstraction/cqrs` offers simple, typed command/query dispatch. This keeps the use case API explicit and replaceable without bringing in heavy frameworks.
- What this buys us here:
  - Swap OCR providers (mock ↔ Yandex) without touching business logic.
  - Easy unit tests via interface mocks for adapters.
  - Clear boundaries: use cases are framework-agnostic; transports and adapters are pluggable.
  - Minimal blast radius for config/secrets; wiring stays in the composition root.
- Data flow: Presentation (gRPC) → Use Case (`extracttext`) → Adapters (`Downloader` → `FileConverter` → `Recognizer`) → back to Presentation.

## Public API (Contracts)

- gRPC service: `ocr.v1.OcrService`.
  - Method: `Process(ParseRequest) -> ParseResponse`.
  - Messages:
    - `ParseRequest { string objectkey = 1; }`
    - `ParseResponse { string text = 1; }`
- grpcurl example (plaintext):

```
grpcurl -plaintext -d '{"objectkey":"folder/file.pdf"}' localhost:8081 ocr.v1.OcrService/Process
```

- Health-check:

```
curl http://localhost:8090/healthz
```

## Configuration (ENV)

- General:
  - `DOC2TEXT_PROVIDER` — OCR provider (`mock` | `yandex`, default `mock`).
- gRPC:
  - `DOC2TEXT_ADDR` — gRPC bind address (default `:8080`).
- HTTP (health):
  - `DOC2TEXT_HTTP_ADDR` — HTTP bind address (default `:8090`).
  - `DOC2TEXT_HTTP_HEALTH_CHECK_PATH` — health path (default `/healthz`).
  - `TG_FORWARDER_API_HEALTH_CHECK_PATH` — optional override for the path.
- Limits:
  - `DOC2TEXT_MAX_FILE_MB` — max file size in MB (default `25`).
  - `DOC2TEXT_MAX_FILES` — max files per request (default `50`).
- Yandex OCR (provide your own secrets):
  - `DOC2TEXT_YC_ENDPOINT`, `DOC2TEXT_YC_API_KEY` or `DOC2TEXT_YC_IAM_TOKEN`, `DOC2TEXT_YC_FOLDER_ID`.
  - `DOC2TEXT_YC_DEFAULT_MODEL`, `DOC2TEXT_YC_LANGUAGES`, `DOC2TEXT_YC_MIN_CONFIDENCE`, `DOC2TEXT_YC_HTTP_TIMEOUT`.
- S3/MinIO (provide your own secrets):
  - `DOC2TEXT_S3_ENDPOINT`, `DOC2TEXT_S3_ACCESS_KEY`, `DOC2TEXT_S3_SECRET_KEY`, `DOC2TEXT_S3_BUCKET`, `DOC2TEXT_S3_USE_SSL`.

## Recreate gRPC

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/presentation/proto/ocr/v1/ocr.proto
```
