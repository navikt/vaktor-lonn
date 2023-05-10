package main

import (
	"context"
	"database/sql"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/navikt/vaktor-lonn/pkg/service"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed pkg/sql/migrations/*.sql
var embedMigrations embed.FS

func onStart(logger *zap.Logger) (service.Handler, error) {
	dbString := getEnv("DB_URL", "postgres://postgres:postgres@127.0.0.1:5432/vaktor")
	azureClientId := os.Getenv("AZURE_APP_CLIENT_ID")
	azureClientSecret := os.Getenv("AZURE_APP_CLIENT_SECRET")
	azureOpenIdTokenEndpoint := os.Getenv("AZURE_OPENID_CONFIG_TOKEN_ENDPOINT")
	minWinTidEndpoint := os.Getenv("MINWINTID_ENDPOINT")
	minWinTidUsername := os.Getenv("MINWINTID_USERNAME")
	minWinTidPassword := os.Getenv("MINWINTID_PASSWORD")
	minWinTidInterval := getEnv("MINWINTID_INTERVAL", "60m")
	vaktorPlanEndpoint := os.Getenv("VAKTOR_PLAN_ENDPOINT")

	handler, err := service.NewHandler(logger, dbString, vaktorPlanEndpoint,
		azureClientId, azureClientSecret, azureOpenIdTokenEndpoint,
		minWinTidUsername, minWinTidPassword, minWinTidEndpoint, minWinTidInterval)
	if err != nil {
		return service.Handler{}, err
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return service.Handler{}, err
	}

	err = goose.Up(handler.DB, "pkg/sql/migrations")

	return handler, err
}

func main() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, err := config.Build()
	if err != nil {
		logger.Error("Problem building logger", zap.Error(err))
		return
	}

	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {

		}
	}(logger)

	logger.Info("Vaktor Lønn starting up...🚀")

	handler, err := onStart(logger)
	if err != nil {
		logger.Error("Problem with onStart", zap.Error(err))
		return
	}

	context, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	handler.Context = context

	go service.Run(handler)

	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			logger.Error("Problem with DB.close", zap.Error(err))
			return
		}
	}(handler.DB)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/period", handler.Period)

	logger.Info("Ready to serve 🙇")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Error("Problem with ListenAndServer", zap.Error(err))
		return
	}
}

func getEnv(key, fallback string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return fallback
}
