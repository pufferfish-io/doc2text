package main

import (
	"context"
	"doc2text/internal/core/abstraction/convert"
	"doc2text/internal/core/abstraction/cqrs"
	"doc2text/internal/core/abstraction/download"
	"doc2text/internal/core/abstraction/logger"
	"doc2text/internal/core/abstraction/recognize"
	"doc2text/internal/core/usecase/extracttext"
	"doc2text/internal/infrastructure/nativeconv"
	"doc2text/internal/infrastructure/s3"
	"doc2text/internal/infrastructure/yocr"
	"doc2text/internal/infrastructure/zaplogger"
	"doc2text/internal/presentation/api"
	config "doc2text/internal/presentation/config"
	"doc2text/internal/presentation/auth"
	ocrv1 "doc2text/internal/presentation/proto/ocr/v1"
	"doc2text/internal/presentation/server/ocr/v1"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
    cfg := mustLoadConfig()

	logger, clean := registerLogger()
	defer clean()

	convertor := registerConverter()

	downloader := registerDownloader(cfg, logger)

	recognizer := registerRecognizer(cfg)

	bus := registerCqrs(CqrsOptions{Logger: logger, Converter: convertor, Downloader: downloader, Recognizer: recognizer})

    grpcSrv := startGRPCServer(cfg, bus, logger)

	httpSrv := startHTTPServer(cfg.HTTP.Addr, cfg.HTTP.HealthCheckPath, logger)

	waitForShutdown(logger, grpcSrv, httpSrv)
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	return cfg
}

func startGRPCServer(cfg *config.Config, bus *cqrs.Bus, l logger.Logger) *grpc.Server {
    lis, err := net.Listen("tcp", cfg.Server.Addr)
    if err != nil {
        l.Error("listen: %v", err)
        os.Exit(1)
    }

    // Attach auth interceptor if OIDC is configured
    var serverOpts []grpc.ServerOption
    if interceptor, err := auth.NewUnaryAuthInterceptor(cfg.OIDC); err != nil {
        l.Error("auth init: %v", err)
        os.Exit(1)
    } else if interceptor != nil {
        serverOpts = append(serverOpts, grpc.UnaryInterceptor(interceptor))
        l.Info("Auth: OIDC interceptor enabled (audience=%s)", cfg.OIDC.Audience)
    } else {
        l.Info("Auth: OIDC not configured; gRPC runs without auth")
    }

    grpcSrv := grpc.NewServer(serverOpts...)
    ocrv1.RegisterOcrServiceServer(grpcSrv, ocr.New(bus))

    l.Info("gRPC listening on %s", cfg.Server.Addr)
    go func() {
        if err := grpcSrv.Serve(lis); err != nil {
            l.Error("serve: %v", err)
        }
    }()
    return grpcSrv
}

func registerLogger() (logger.Logger, func()) {
	logger, cleanup := zaplogger.NewZapLogger()
	return logger, cleanup
}

func registerConverter() convert.FileConverter {
	converter := nativeconv.NewFileConverter()
	return converter
}
func registerDownloader(cfg *config.Config, l logger.Logger) download.Downloader {
	downloader, err := s3.NewDownloader(s3.Config{
		Endpoint:  cfg.S3.Endpoint,
		AccessKey: cfg.S3.AccessKey,
		SecretKey: cfg.S3.SecretKey,
		Bucket:    cfg.S3.Bucket,
		UseSSL:    cfg.S3.UseSSL,
	})
	if err != nil {
		l.Error("s3.NewDownloader: %v", err)
		os.Exit(1)
	}
	return downloader
}
func registerRecognizer(cfg *config.Config) recognize.Recognizer {
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

func registerCqrs(o CqrsOptions) *cqrs.Bus {
	extractH := extracttext.NewHandler(o.Converter, o.Downloader, o.Logger, o.Recognizer)
	bus := cqrs.NewBus()
	cqrs.RegisterQuery(bus, extractH)
	return bus
}

func startHTTPServer(httpAddr string, healthPath string, l logger.Logger) *http.Server {
	httpMux := api.NewRouter(api.Options{HealthCheckPath: healthPath})
	srv := &http.Server{Addr: httpAddr, Handler: httpMux}
	l.Info("HTTP listening on %s (health: %s)", httpAddr, healthPath)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("http serve: %v", err)
		}
	}()
	return srv
}

func waitForShutdown(l logger.Logger, grpcSrv *grpc.Server, httpSrv *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	l.Info("Service started; press Ctrl+C to stop")
	<-c
	l.Info("Shutdown signal received, exiting")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if httpSrv != nil {
		if err := httpSrv.Shutdown(ctx); err != nil {
			l.Error("http shutdown: %v", err)
		}
	}

	if grpcSrv != nil {
		done := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-ctx.Done():
			grpcSrv.Stop()
		}
	}
}
