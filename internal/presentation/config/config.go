package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Core struct {
	Provider string `env:"PROVIDER" envDefault:"mock"` // mock | yandex
}

type Server struct {
	Addr string `env:"ADDR" envDefault:":8080"` // gRPC bind
}

type Limits struct {
	MaxFileMB int `env:"MAX_FILE_MB" envDefault:"25"`
	MaxFiles  int `env:"MAX_FILES"   envDefault:"50"`
}

type Yandex struct {
	APIKey        string        `env:"API_KEY"`   // DOC2TEXT_YC_API_KEY
	IAMToken      string        `env:"IAM_TOKEN"` // DOC2TEXT_YC_IAM_TOKEN (пока не используется в yocr)
	FolderID      string        `env:"FOLDER_ID"` // DOC2TEXT_YC_FOLDER_ID
	Endpoint      string        `env:"ENDPOINT"      envDefault:"https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze"`
	Model         string        `env:"DEFAULT_MODEL" envDefault:"page"` // page | handwritten | markdown | math-markdown
	MinConfidence float64       `env:"MIN_CONFIDENCE" envDefault:"0.6"`
	HTTPTimeout   time.Duration `env:"HTTP_TIMEOUT"  envDefault:"15s"`

	// 🔧 added: список языков через запятую, например: "ru,en"
	Languages []string `env:"LANGUAGES" envSeparator:"," envDefault:"ru,en"`
}

// 🔧 added: конфиг для MinIO/S3
type S3 struct {
	Endpoint  string `env:"ENDPOINT,required"`   // DOC2TEXT_S3_ENDPOINT (например: "127.0.0.1:9000")
	AccessKey string `env:"ACCESS_KEY,required"` // DOC2TEXT_S3_ACCESS_KEY
	SecretKey string `env:"SECRET_KEY,required"` // DOC2TEXT_S3_SECRET_KEY
	Bucket    string `env:"BUCKET,required"`     // DOC2TEXT_S3_BUCKET
	UseSSL    bool   `env:"USE_SSL" envDefault:"false"`
}

type Config struct {
	Core   Core   `envPrefix:"DOC2TEXT_"`    // → DOC2TEXT_PROVIDER
	Server Server `envPrefix:"DOC2TEXT_"`    // → DOC2TEXT_ADDR
	Limits Limits `envPrefix:"DOC2TEXT_"`    // → DOC2TEXT_MAX_FILE_MB, ...
	Yandex Yandex `envPrefix:"DOC2TEXT_YC_"` // → DOC2TEXT_YC_*
	S3     S3     `envPrefix:"DOC2TEXT_S3_"` // 🔧 added → DOC2TEXT_S3_*
}

func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}

	// Лёгкая валидация для провайдера Yandex
	if c.Core.Provider == "yandex" {
		if c.Yandex.FolderID == "" {
			return nil, fmt.Errorf("DOC2TEXT_YC_FOLDER_ID is required for provider=yandex")
		}
		if c.Yandex.APIKey == "" && c.Yandex.IAMToken == "" {
			return nil, fmt.Errorf("either DOC2TEXT_YC_API_KEY or DOC2TEXT_YC_IAM_TOKEN is required for provider=yandex")
		}
	}
	return &c, nil
}
