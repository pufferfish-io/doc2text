package config

import (
    "fmt"
    "time"

    "github.com/caarlos0/env/v11"
    "github.com/go-playground/validator/v10"
)

type Core struct {
    Provider string `env:"PROVIDER" envDefault:"mock" validate:"oneof=mock yandex"`
}

type Server struct {
    Addr string `env:"ADDR" envDefault:":8080" validate:"required"`
}

type HTTP struct {
    Addr            string `env:"ADDR"               envDefault:":8090" validate:"required"`
    HealthCheckPath string `env:"HEALTH_CHECK_PATH" envDefault:"/healthz" validate:"required"`
}

type Limits struct {
    MaxFileMB int `env:"MAX_FILE_MB" envDefault:"25" validate:"gte=1"`
    MaxFiles  int `env:"MAX_FILES"   envDefault:"50" validate:"gte=1"`
}

type Yandex struct {
    APIKey        string        `env:"API_KEY"`
    IAMToken      string        `env:"IAM_TOKEN"`
    FolderID      string        `env:"FOLDER_ID"`
    Endpoint      string        `env:"ENDPOINT"      envDefault:"https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze"`
    Model         string        `env:"DEFAULT_MODEL" envDefault:"page"`
    MinConfidence float64       `env:"MIN_CONFIDENCE" envDefault:"0.6" validate:"gte=0,lte=1"`
    HTTPTimeout   time.Duration `env:"HTTP_TIMEOUT"  envDefault:"15s" validate:"gt=0"`
    Languages     []string      `env:"LANGUAGES" envSeparator:"," envDefault:"ru,en"`
}

type S3 struct {
    Endpoint  string `env:"ENDPOINT,required" validate:"required"`
    AccessKey string `env:"ACCESS_KEY,required" validate:"required"`
    SecretKey string `env:"SECRET_KEY,required" validate:"required"`
    Bucket    string `env:"BUCKET,required" validate:"required"`
    UseSSL    bool   `env:"USE_SSL" envDefault:"false"`
}

type OIDC struct {
    Issuer      string `env:"ISSUER"      validate:"required_with=JWKSURL Audience ExpectedAzp"`
    JWKSURL     string `env:"JWKS_URL"    validate:"required_with=Issuer Audience ExpectedAzp,url"`
    Audience    string `env:"AUDIENCE"    validate:"required_with=Issuer JWKSURL ExpectedAzp"`
    ExpectedAzp string `env:"EXPECTED_AZP"`
}

type Config struct {
    Core   Core   `envPrefix:"DOC2TEXT_"`
    Server Server `envPrefix:"DOC2TEXT_"`
    HTTP   HTTP   `envPrefix:"DOC2TEXT_HTTP_"`
    Limits Limits `envPrefix:"DOC2TEXT_"`
    Yandex Yandex `envPrefix:"DOC2TEXT_YC_"`
    S3     S3     `envPrefix:"DOC2TEXT_S3_"`
    OIDC   OIDC   `envPrefix:"DOC2TEXT_OIDC_"`
}

func Load() (*Config, error) {
    var c Config
    if err := env.Parse(&c); err != nil {
        return nil, fmt.Errorf("env parse: %w", err)
    }
    v := validator.New()
    if err := v.Struct(c); err != nil {
        return nil, fmt.Errorf("config validate: %w", err)
    }

    return &c, nil
}
