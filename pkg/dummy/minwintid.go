package dummy

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strings"
)

func generateWorkhours() string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", rand.Intn(4)+6, rand.Intn(60), rand.Intn(4)+14, rand.Intn(60))
}

func GetMinWinTid(plan models.Vaktplan) map[string][]string {
	log.Printf("Creating fake timesheet for %s", plan.Ident)
	timesheet := map[string][]string{}
	for day, _ := range plan.Schedule {
		// TODO: Bruk samme datoformat som Vaktor Plan
		splits := strings.Split(day, "-")
		timesheet[fmt.Sprintf("%v.%v.%v", splits[2], splits[1], splits[1])] = []string{generateWorkhours()}
	}

	return timesheet
}
