package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"

	"os"
	"time"
)

var ServiceName = "clickhouse-protocol-proxy"
var Version = "1.0"

type Config struct {
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
	Target struct {
		Hosts    []string       `yaml:"hosts"`
		Settings map[string]any `yaml:"settings"`

		MaxConnectionsPerUser int           `yaml:"maxConnectionPerUser"`
		MaxConnectionLifetime time.Duration `yaml:"maxConnectionLifetime"`
		DialTimeout           time.Duration `yaml:"dialTimeout"`
		ReadTimeout           time.Duration `yaml:"readTimeout"`

		Debug bool `yaml:"debug"`
	} `yaml:"target"`
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
}

func New(path string) (*Config, *zap.Logger) {
	cfgFileContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(cfgFileContent, cfg)
	if err != nil {
		panic(err)
	}

	zcfg := zap.NewProductionConfig()
	zcfg.Level, _ = zap.ParseAtomicLevel(cfg.Log.Level)
	logger := zap.Must(zcfg.Build()).With(zap.Field{Type: zapcore.StringType, Key: "service", String: ServiceName}, zap.Field{Type: zapcore.StringType, Key: "version", String: Version})

	return cfg, logger
}
