package utils

import (
	"os"

	"go.uber.org/zap"
)

func SetUpLogger() *zap.Logger {
	switch os.Getenv("APP_ENV") {
	case "production":
		log, _ := zap.NewProduction()
		// log.Sync()
		return log
	default:
		log, _ := zap.NewDevelopment()
		return log
	}
}
