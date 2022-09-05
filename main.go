package main

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Print("Vaktor Lønn starting up...")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/period", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var plan models.Plan
			err := json.NewDecoder(r.Body).Decode(&plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				log.Err(err)
				return
			}
			// TODO: Skal vi validere input?
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

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Err(err)
		return
	}
}
