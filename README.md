# doc2text

## Что делает

1. Включает gRPC-сервис `OcrService` и принимает запросы с `objectKey`; скачивает файл из S3, конвертирует в `Base64` и отправляет его на Yandex Vision OCR.
2. Формирует `ExtractTextResponse` с распознанным текстом, метаданными и логирует все шаги через `internal/logger`.
3. Параллельно запускает HTTP-сервер с `/healthz`, чтобы оркестраторы могли проверять готовность, и слушает настраиваемый gRPC-порт.
4. При наличии OIDC-конфигурации оборачивает gRPC-методы перехватчиком (`auth.NewUnaryAuthInterceptor`) и требует JWT с заданными `audience`/`issuer`.

## Запуск

1. Настройте переменные окружения (см. следующий раздел) — можно положить их в `.env` и загрузить `set -a && source .env && set +a`.
2. Соберите и запустите из корня:
   ```bash
   go run ./cmd/doc2text
   ```
3. Или создайте Docker-образ и запустите его, передав нужные переменные:
   ```bash
   docker build -t doc2text .
   docker run --rm -e ... doc2text
   ```

## Переменные окружения

Все переменные обязательны, кроме SASL/OIDC/Yandex IAM, которые используют только в защищённых окружениях.

- `G_RPC_SERVER_DOC2TEXT_ADDR` — адрес, на котором слушает gRPC (`:8080` по умолчанию).
- `HTTP_SERVER_DOC2TEXT_ADDR` — адрес HTTP/health-сервера (`:8090` по умолчанию).
- `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` — параметры S3-совместимого хранилища, из которого скачиваются файлы.
- `S3_USE_SSL` — `true/false`, включает HTTPS при работе с S3 (по умолчанию `false`).
- `YC_API_KEY`, `YC_IAM_TOKEN`, `YC_FOLDER_ID` — учётные данные Yandex Vision (любая из двух комбинаций с IAM работает).
- `YC_ENDPOINT` — URL Yandex Vision API, по умолчанию `https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze`.
- `YC_DEFAULT_MODEL` — модель Vision (`page` по умолчанию).
- `YC_MIN_CONFIDENCE` — минимальная граница доверия `0.0…1.0` (по умолчанию `0.6`).
- `YC_HTTP_TIMEOUT` — таймаут HTTP-запросов к Vision (`15s` по умолчанию).
- `YC_LANGUAGES` — список языков (записывается через запятую, по умолчанию `ru,en`).
- `OIDC_DOC2TEXT_ISSUER`, `OIDC_DOC2TEXT_JWKS_URL`, `OIDC_DOC2TEXT_AUDIENCE`, `OIDC_DOC2TEXT_EXPECTED_AZP` — при задании `issuer`/`jwks_url`/`audience` сервис включает проверку JWT и требует валидный токен на gRPC-запросы.

## Примечания

- Все детали архитектуры описаны в `docs/system-design/README.md`, а примеры вызовов можно найти в `docs/usage/README.md`.
- Код, который скачивает файл, находится в `internal/infrastructure/s3`, распознавание — в `internal/infrastructure/yocr`, обёртки — в `internal/core`.
- `auth.NewUnaryAuthInterceptor` отключён, если OIDC-данные не заданы; в этом случае gRPC работает без авторизации.
- HTTP-сервер отдаёт `200 OK` на `/healthz` и не обрабатывает другие маршруты.
