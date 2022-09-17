package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/calculator"
	"github.com/navikt/vaktor-lonn/pkg/dummy"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h Handler) Nudge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// TODO: Trigger sjekk av MinWinTid og hva vi har i databasen.
	beredskapsvakter, err := h.Queries.ListBeredskapsvakter(context.TODO())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
		log.Error().Msg(err.Error())
		return
	}

	for _, beredskapsvakt := range beredskapsvakter {
		var vaktplan models.Vaktplan
		err := json.Unmarshal(beredskapsvakt.Plan, &vaktplan)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Error().Msg(err.Error())
			return
		}

		// TODO: Bytt med en korrekt implementasjon av kommunikasjon med MinWinTid
		minWinTid := dummy.GetMinWinTid(vaktplan)

		log.Printf("Calculating salary for %s", beredskapsvakt.Ident)
		// TODO: Lage transaksjonsliste
		// TODO: Bytt ut med en go routine, da vi ikke skal svare med lønn her
		report, err := calculator.GuarddutySalary(vaktplan, minWinTid)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Error().Msg(err.Error())
			return
		}
		err = json.NewEncoder(w).Encode(report)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusBadRequest)
			log.Error().Msg(err.Error())
			return
		}
	}
}
