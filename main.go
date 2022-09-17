package main

import (
	"database/sql"
	"embed"
	"github.com/navikt/vaktor-lonn/pkg/endpoints"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"os"
)

//go:embed pkg/sql/migrations/*.sql
var embedMigrations embed.FS

func onStart(logger *zap.Logger) (endpoints.Handler, error) {
	dbString := getEnv("NAIS_DATABASE_NADA_BACKEND_NADA_URL", "postgres://postgres:postgres@127.0.0.1:5432/vaktor")
	handler, err := endpoints.NewHandler(logger, dbString)
	if err != nil {
		return handler, err
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return handler, err
	}

	err = goose.Up(handler.DB, ".")

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
	http.HandleFunc("/nudge", handler.Nudge)
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
