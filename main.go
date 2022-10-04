package main

import (
	"database/sql"
	"embed"
	"github.com/navikt/vaktor-lonn/pkg/endpoints"
	"net/http"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

//go:embed pkg/sql/migrations/*.sql
var embedMigrations embed.FS

func onStart(logger *zap.Logger) (endpoints.Handler, error) {
	dbString := getEnv("DB_URL", "postgres://postgres:postgres@127.0.0.1:5432/vaktor")
	minWinTidEndpoint := os.Getenv("MINWINTID_ENDPOINT")
	minWinTidUsername := os.Getenv("MINWINTID_USERNAME")
	minWinTidPassword := os.Getenv("MINWINTID_PASSWORD")

	handler, err := endpoints.NewHandler(logger, dbString,
		minWinTidUsername, minWinTidPassword, minWinTidEndpoint)
	if err != nil {
		return endpoints.Handler{}, err
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return endpoints.Handler{}, err
	}

	err = goose.Up(handler.DB, "pkg/sql/migrations")

	return handler, err
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Vaktor Lønn starting up...🚀")

	handler, err := onStart(logger)
	if err != nil {
		logger.Error("Problem with onStart", zap.Error(err))
		return
	}

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
