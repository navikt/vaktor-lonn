package dummy

import (
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"math/rand"
)

func GetSatserFromAgresso() map[string]float64 {
	return map[string]float64{
		"lørsøn":  55,
		"0620":    10,
		"2006":    20,
		"utvidet": 15,
	}
}

func GetSalary(ident string) decimal.Decimal {
	log.Printf("Creating fake salary for %s", ident)
	return decimal.NewFromInt(int64(rand.Intn(700_000) + 300_000))
}
