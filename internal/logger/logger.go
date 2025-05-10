package logger

import (
	"Brocker-pet-project/internal/config"
	"go.uber.org/zap"
)

func InitLogger(cfg *config.Config) (*zap.Logger, error) {

	switch cfg.Env {
	case "dev":
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		return cfg.Build()
	case "prod":
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		return cfg.Build()
	default: //local
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		return cfg.Build()
	}

}
