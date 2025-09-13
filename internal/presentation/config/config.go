package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Core struct {
	Provider string `env:"PROVIDER" envDefault:"mock"`
}

type Server struct {
    Addr string `env:"ADDR" envDefault:":8080"`
}

type HTTP struct {
    Addr            string `env:"ADDR"               envDefault:":8090"`
    HealthCheckPath string `env:"HEALTH_CHECK_PATH" envDefault:"/healthz"`
}

type Limits struct {
	MaxFileMB int `env:"MAX_FILE_MB" envDefault:"25"`
	MaxFiles  int `env:"MAX_FILES"   envDefault:"50"`
}

type Yandex struct {
	APIKey        string        `env:"API_KEY"`
	IAMToken      string        `env:"IAM_TOKEN"`
	FolderID      string        `env:"FOLDER_ID"`
	Endpoint      string        `env:"ENDPOINT"      envDefault:"https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze"`
	Model         string        `env:"DEFAULT_MODEL" envDefault:"page"`
	MinConfidence float64       `env:"MIN_CONFIDENCE" envDefault:"0.6"`
	HTTPTimeout   time.Duration `env:"HTTP_TIMEOUT"  envDefault:"15s"`
	Languages []string `env:"LANGUAGES" envSeparator:"," envDefault:"ru,en"`
}

type S3 struct {
	Endpoint  string `env:"ENDPOINT,required"`
	AccessKey string `env:"ACCESS_KEY,required"`
	SecretKey string `env:"SECRET_KEY,required"`
	Bucket    string `env:"BUCKET,required"`
	UseSSL    bool   `env:"USE_SSL" envDefault:"false"`
}

type Config struct {
    Core   Core   `envPrefix:"DOC2TEXT_"`
    Server Server `envPrefix:"DOC2TEXT_"`
    HTTP   HTTP   `envPrefix:"DOC2TEXT_HTTP_"`
    Limits Limits `envPrefix:"DOC2TEXT_"`
    Yandex Yandex `envPrefix:"DOC2TEXT_YC_"`
    S3     S3     `envPrefix:"DOC2TEXT_S3_"`
}

func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}

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
