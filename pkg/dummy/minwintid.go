package dummy

import (
	"fmt"
	"github.com/navikt/vaktor-lonn/pkg/models"
	"github.com/shopspring/decimal"
	"math/rand"
	"strings"
)

func getSatser() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"lørsøn":  decimal.NewFromInt(55),
		"0620":    decimal.NewFromInt(10),
		"2006":    decimal.NewFromInt(20),
		"utvidet": decimal.NewFromInt(15),
	}
}

func getSalary() decimal.Decimal {
	return decimal.NewFromInt(int64(rand.Intn(700_000) + 300_000))
}

func generateWorkhours() string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", rand.Intn(4)+6, rand.Intn(60), rand.Intn(4)+14, rand.Intn(60))
}

func GetMinWinTid(bearerToken string, plan models.Vaktplan) models.MinWinTid {
	minWinTid := models.MinWinTid{
		Salary:    getSalary(),
		Satser:    getSatser(),
		Timesheet: map[string][]string{},
	}

	for day, _ := range plan.Schedule {
		// TODO: Bruk samme datoformat som Vaktor Plan
		splits := strings.Split(day, "-")
		properDay := fmt.Sprintf("%v.%v.%v", splits[2], splits[1], splits[1])
		minWinTid.Timesheet[properDay] = []string{generateWorkhours()}
	}

	return minWinTid
}
