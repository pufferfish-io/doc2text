package main

import (
	"doc2text/internal/core/abstraction/convert"
	"doc2text/internal/core/abstraction/cqrs"
	"doc2text/internal/core/abstraction/download"
	"doc2text/internal/core/abstraction/logger"
	"doc2text/internal/core/abstraction/recognize"
	"doc2text/internal/core/usecase/extracttext"
	"doc2text/internal/infrastructure/adapter/nativeconv"
	"doc2text/internal/infrastructure/adapter/s3"
	"doc2text/internal/infrastructure/adapter/yocr"
	"doc2text/internal/infrastructure/adapter/zaplogger"
	config "doc2text/internal/presentation/config"
	ocrv1 "doc2text/internal/presentation/proto/ocr/v1"
	"doc2text/internal/presentation/server/ocr/v1"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	cfg := mustLoadConfig()

	logger, clean := RegisterLogger()
	defer clean()

	convertor := RegisterConverter()

	downloader := RegisterDownloader(cfg)

	recognizer := RegisterRecognizer(cfg)

	cqrsOpt := CqrsOptions{
		Logger:     logger,
		Converter:  convertor,
		Downloader: downloader,
		Recognizer: recognizer,
	}
	bus := RegisterCqrs(cqrsOpt)

	funcMustRegisterGRpc(cfg.Server.Addr, bus)
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	return cfg
}

func funcMustRegisterGRpc(port string, bus *cqrs.Bus) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcSrv := grpc.NewServer()

	ocrv1.RegisterOcrServiceServer(grpcSrv, ocr.New(bus))

	log.Printf("gRPC listening on :%s", port)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func RegisterLogger() (logger.Logger, func()) {
	logger, cleanup := zaplogger.NewZapLogger()
	return logger, cleanup
}

func RegisterConverter() convert.FileConverter {
	converter := nativeconv.NewFileConverter()
	return converter
}
func RegisterDownloader(cfg *config.Config) download.Downloader {
	downloader, err := s3.NewDownloader(s3.Config{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		UseSSL:    cfg.S3.UseSSL,
	})
	if err != nil {
		log.Fatalf("s3.NewDownloader: %v", err)
	}
	return downloader
}
func RegisterRecognizer(cfg *config.Config) recognize.Recognizer {
	recognizer := yocr.New(yocr.Options{
		OcrEndpoint: cfg.Yandex.Endpoint,
		ApiKey:      cfg.Yandex.APIKey,
		FolderID:    cfg.Yandex.FolderID,
		Model:       cfg.Yandex.Model,
		Languages:   cfg.Yandex.Languages,
	})
	return recognizer
}

type CqrsOptions struct {
	Logger     logger.Logger
	Converter  convert.FileConverter
	Downloader download.Downloader
	Recognizer recognize.Recognizer
}

func RegisterCqrs(o CqrsOptions) *cqrs.Bus {
	extractH := extracttext.NewHandler(o.Converter, o.Downloader, o.Logger, o.Recognizer)
	bus := cqrs.NewBus()
	cqrs.RegisterQuery(bus, extractH)

	return bus
}
