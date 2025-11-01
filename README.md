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
grpcurl -plaintext -d '{"objectkey":"folder/file.pdf"}' localhost:50051 ocr.v1.OcrService/Process
```

- Go client example:

```
go run ./examples/grpc-client \
  | OIDC_TOKEN_URL=https://auth.pufferfish.ru/realms/pufferfish/protocol/openid-connect/token \
    OIDC_CLIENT_ID=message-responder-ocr \
    OIDC_CLIENT_SECRET=*** \
    OIDC_SCOPE=openid \
    GRPC_TARGET=localhost:50052 \
    OBJECT_KEY=folder/file.pdf
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
- Limits:
  - `DOC2TEXT_MAX_FILE_MB` — max file size in MB (default `25`).
  - `DOC2TEXT_MAX_FILES` — max files per request (default `50`).
- Yandex OCR (provide your own secrets):
  - `DOC2TEXT_YC_ENDPOINT`, `DOC2TEXT_YC_API_KEY` or `DOC2TEXT_YC_IAM_TOKEN`, `DOC2TEXT_YC_FOLDER_ID`.
  - `DOC2TEXT_YC_DEFAULT_MODEL`, `DOC2TEXT_YC_LANGUAGES`, `DOC2TEXT_YC_MIN_CONFIDENCE`, `DOC2TEXT_YC_HTTP_TIMEOUT`.
- S3/MinIO (provide your own secrets):
  - `DOC2TEXT_S3_ENDPOINT`, `DOC2TEXT_S3_ACCESS_KEY`, `DOC2TEXT_S3_SECRET_KEY`, `DOC2TEXT_S3_BUCKET`, `DOC2TEXT_S3_USE_SSL`.

### Auth (OIDC / Keycloak)

- When these vars are set, gRPC enforces JWT validation for every call (prefix `DOC2TEXT_OIDC_`):
  - `DOC2TEXT_OIDC_ISSUER` — `https://auth.pufferfish.ru/realms/pufferfish`
  - `DOC2TEXT_OIDC_JWKS_URL` — `https://auth.pufferfish.ru/realms/pufferfish/protocol/openid-connect/certs`
  - `DOC2TEXT_OIDC_AUDIENCE` — `doc2text`
  - `DOC2TEXT_OIDC_EXPECTED_AZP` — `message-responder-ocr` (optional, but recommended)

- Expected token (client_credentials via Keycloak, simplified):

```
{
  "iss": "https://auth.pufferfish.ru/realms/pufferfish",
  "aud": ["doc2text"],
  "azp": "message-responder-ocr",
  "client_id": "message-responder-ocr",
  "scope": "email profile"
}
```

- grpcurl example with Bearer token:

```
grpcurl -plaintext \
  -H "authorization: Bearer $ACCESS_TOKEN" \
  -d '{"objectkey":"folder/file.pdf"}' \
localhost:50051 ocr.v1.OcrService/Process
```

Note: for server configuration, use the `DOC2TEXT_OIDC_*` variables shown above. Client credentials envs (`OIDC_TOKEN_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_SCOPE`) are used only by the example client to obtain the token.

## Recreate gRPC

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/presentation/proto/ocr/v1/ocr.proto
```

## Command Guide

### Run with exported .env (one‑liner)

Exports all variables from `.env` into the current shell and runs the service.

```
export $(cat .env | xargs) && go run ./cmd/doc2text
```

### Run with `source` (safer for complex values)

Loads `.env` preserving quotes and special characters, then runs the service.

```
set -a && source .env && set +a && go run ./cmd/doc2text
```

### Fetch/clean module deps

Resolves dependencies and prunes unused ones.

```
go mod tidy
```

### Verbose build (diagnostics)

Builds the binary with verbose and command tracing. Removes old binary after build to keep the tree clean.

```
go build -v -x ./cmd/doc2text && rm -f doc2text
```

### Docker build (Buildx)

Builds the image with detailed progress logs and without cache.

```
docker buildx build --no-cache --progress=plain .
```

### Create and push tag

Cuts a release tag and pushes it to remote.

```
git tag v0.0.1
git push origin v0.0.1
```

### Manage tags

List all tags, delete a tag locally and remotely, verify deletion.

```
git tag -l
git tag -d vX.Y.Z
git push --delete origin vX.Y.Z
git ls-remote --tags origin | grep 'refs/tags/vX.Y.Z$'
```
