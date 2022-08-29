package main

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	fmt.Println("Vaktor Lønn starting up...")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/period", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var plan models.Plan
			err := json.NewDecoder(r.Body).Decode(&plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				return
			}
			report, err := calculator.GuarddutySalary(plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				return
			}
			err = json.NewEncoder(w).Encode(report)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
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
		return
	}
}
