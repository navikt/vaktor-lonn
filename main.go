package main

import (
	"database/sql"
	"embed"
	"net/http"
	"os"

	"github.com/navikt/vaktor-lonn/pkg/endpoints"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:embed pkg/sql/migrations/*.sql
var embedMigrations embed.FS

func onStart() (endpoints.Handler, error) {
	dbString := getEnv("NAIS_DATABASE_NADA_BACKEND_NADA_URL", "postgres://postgres:postgres@127.0.0.1:5432/vaktor")
	handler, err := endpoints.NewHandler(dbString)
	if err != nil {
		return handler, err
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return handler, err
	}

	err = goose.Up(handler.DB, "migrations")

	return handler, err
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Print("Vaktor Lønn starting up...")

	handler, err := onStart()
	if err != nil {
		log.Err(err).Msg("Problem with onStart")
		return
	}

	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			log.Err(err).Msg("Problem with DB.close")
			return
		}
	}(handler.DB)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/nudge", handler.Nudge)
	http.HandleFunc("/period", handler.Period)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Err(err).Msg("Problem with ListenAndServe")
		return
	}
}

func getEnv(key, fallback string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return fallback
}
