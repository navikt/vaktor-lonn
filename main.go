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

	"github.com/navikt/vaktor-lonn/pkg/auth"
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
	azureClientID := os.Getenv("AZURE_APP_CLIENT_ID")
	azureClientSecret := os.Getenv("AZURE_APP_CLIENT_SECRET")
	azureOpenIDTokenEndpoint := os.Getenv("AZURE_OPENID_CONFIG_TOKEN_ENDPOINT")
	minWinTidORDSEndpoint := os.Getenv("MINWINTID_ORDS_ENDPOINT")
	minWinTidEndpoint := os.Getenv("MINWINTID_ENDPOINT")
	minWinTidClientID := os.Getenv("MINWINTID_CLIENTID")
	minWinTidSecret := os.Getenv("MINWINTID_SECRET")
	minWinTidInterval := getEnv("MINWINTID_INTERVAL", "60m")
	vaktorPlanEndpoint := os.Getenv("VAKTOR_PLAN_ENDPOINT")

	minWinTidTicketInterval, err := time.ParseDuration(minWinTidInterval)
	if err != nil {
		return service.Handler{}, err
	}

	minWinTidConfig := service.MinWinTidConfig{
		BearerClient:   auth.NewWithBasicAuth(minWinTidClientID, minWinTidSecret, minWinTidORDSEndpoint),
		Endpoint:       minWinTidEndpoint,
		TickerInterval: minWinTidTicketInterval,
	}

	handler, err := service.NewHandler(logger, dbString, azureClientID, azureClientSecret, azureOpenIDTokenEndpoint, vaktorPlanEndpoint, minWinTidConfig)
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
		err = logger.Sync()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}(logger)

	logger.Info("Vaktor Lønn starting up...🚀")

	handler, err := onStart(logger)
	if err != nil {
		logger.Error("Problem with onStart", zap.Error(err))
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	handler.Context = ctx

	go service.Run(handler)

	defer func(DB *sql.DB) {
		err = DB.Close()
		if err != nil {
			logger.Error("Problem with DB.close", zap.Error(err))
			return
		}
	}(handler.DB)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/period", handler.Period)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("Ready to serve 🙇")
	err = srv.ListenAndServe()
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
