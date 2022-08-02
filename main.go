package main

import (
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"net/http"
)

func main() {
	http.HandleFunc("/period", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var plan models.Plan
			err := json.NewDecoder(r.Body).Decode(&plan)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
				return
			}
			money, err := calculator.GuarddutySalary(plan)
			_, err = fmt.Fprintf(w, "{\"earnings\": \"%f\"}", money)
			if err != nil {
				return
			}
			return
		}
		_, err := fmt.Fprintln(w, "Hello, we only support POST")
		if err != nil {
			return
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
