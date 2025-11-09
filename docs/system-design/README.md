# Системный дизайн и устройство

Общее
- doc2text — сервис, который принимает ключ объекта в S3, скачивает файл, конвертирует его в Base64 и отправляет в Yandex OCR. Результат (распознанный текст) возвращается по gRPC.
- Присутствует простой HTTP‑эндпоинт для health‑check.

Слои (Clean Architecture)
- Контракты домена: `internal/core/abstraction/*`
  - `convert.FileConverter` — конвертация файла в Base64
  - `download.Downloader` — скачивание и метаданные из S3/MinIO
  - `recognize.Recognizer` — распознавание текста (Yandex OCR)
  - `logger.Logger` — логирование
  - Лёгкий CQRS‑шин: `internal/core/abstraction/cqrs`
- Юзкейс: `internal/core/usecase/extracttext`
  - Оркестрирует скачивание → Base64 → распознавание
- Инфраструктура (адаптеры): `internal/infrastructure/*`
  - `s3` — MinIO клиент для загрузки файлов
  - `nativeconv` — Base64‑конвертация
  - `yocr` — клиент Yandex OCR API
  - `zaplogger` — логгер на базе `zap`
- Презентация: `internal/presentation/*`
  - gRPC сервис: `server/ocr/v1` + сгенерированные `proto/ocr/v1`
  - HTTP health‑router: `api/router.go`
  - Конфигурация: `config`
  - OIDC‑интерцептор: `auth/interceptor.go`
- Composition root: `cmd/doc2text/main.go`
  - Сборка зависимостей и запуск серверов

Поток запроса
1. gRPC: `ocr.v1.OcrService/Process(objectkey)`
2. `extracttext` запрашивает:
   - `download.GetInfo` → MIME‑тип
   - `download.GetFile` → байты файла
3. `convert.ToBase64` — потоковая Base64‑кодировка (чанки ~64KB)
4. `recognize.Recognize` — запрос к Yandex OCR и парсинг ответа
5. Возврат `text` в ответе gRPC

gRPC API (кратко)
- Сервис: `ocr.v1.OcrService`
- Метод: `Process(ParseRequest{objectkey}) -> ParseResponse{text}`

Аутентификация (OIDC)
- Если заданы переменные `OIDC_DOC2TEXT_*`, включается верификация JWT в gRPC через unary‑interceptor.
- Проверяются issuer, audience, подпись и (опционально) `azp`.
- При отсутствии настроек — сервис работает без аутентификации.

Конфигурация (ENV, префиксы)
- gRPC: `G_RPC_SERVER_DOC2TEXT_...` → `ADDR`
- HTTP: `HTTP_SERVER_DOC2TEXT_...` → `ADDR`
- S3: `S3_...` → `ENDPOINT`, `ACCESS_KEY`, `SECRET_KEY`, `BUCKET`, `USE_SSL`
- Yandex OCR: `YC_...` → `API_KEY`, `FOLDER_ID`, `ENDPOINT`, `DEFAULT_MODEL`, `LANGUAGES`, `MIN_CONFIDENCE`, `HTTP_TIMEOUT`
- OIDC: `OIDC_DOC2TEXT_...` → `ISSUER`, `JWKS_URL`, `AUDIENCE`, `EXPECTED_AZP`

Заметки по реализации
- HTTP‑клиент Yandex OCR использует таймаут 30s.
- MIME‑тип берётся из метаданных объекта; при пустом — определяется по расширению, затем дефолт `application/octet-stream`.
- Health‑эндпоинт по умолчанию: `GET /healthz`.

Стартовые точки кода
- Конфигурация: internal/presentation/config/config.go:1
- Сборка и запуск: cmd/doc2text/main.go:1
- gRPC сервис: internal/presentation/server/ocr/v1/service.go:1
- Юзкейс: internal/core/usecase/extracttext/handler.go:1
- Адаптеры: internal/infrastructure/{s3,yocr,nativeconv}/service.go:1
