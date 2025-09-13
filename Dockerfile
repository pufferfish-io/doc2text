FROM golang:1.23.12 AS builder
WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION
ARG COMMIT
ARG BUILT_AT

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o app ./cmd/doc2text

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

COPY --from=builder /app/app /app/app

EXPOSE 8080 8090
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
