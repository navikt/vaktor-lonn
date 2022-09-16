package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	gensql "github.com/navikt/vaktor-lonn/pkg/sql/gen"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

//go:embed pkg/sql/migrations/*.sql
var embedMigrations embed.FS

func setupDB() (*sql.DB, error) {
	dbString := getEnv("NAIS_DATABASE_NADA_BACKEND_NADA_URL", "postgres://postgres:postgres@127.0.0.1:5432/vaktor")
	var db *sql.DB
	db, err := sql.Open("postgres", dbString)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return nil, err
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	log.Print("Vaktor Lønn starting up...")
	db, err := setupDB()
	if err != nil {
		log.Err(err)
		return
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/period", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var plan models.Vaktplan
			err := json.NewDecoder(r.Body).Decode(&plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				log.Err(err)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				log.Err(err)
				return
			}

			// TODO: Skal vi validere input?
			queries := gensql.New(db)
			err = queries.CreatePlan(context.TODO(), gensql.CreatePlanParams{
				ID:    plan.ID,
				Ident: plan.Ident,
				Plan:  body,
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
				log.Err(err)
				return
			}

			log.Printf("Calculating salary for %s", plan.Ident)
			report, err := calculator.GuarddutySalary(plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				log.Err(err)
				return
			}
			err = json.NewEncoder(w).Encode(report)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				log.Err(err)
				return
			}
			return
		}
		_, err := fmt.Fprintln(w, "Hello, we only support POST")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			return
		}
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Err(err)
		return
	}
}

func getEnv(key, fallback string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return fallback
}
