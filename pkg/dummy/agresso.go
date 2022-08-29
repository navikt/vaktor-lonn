package dummy

import (
	"fmt"
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

func GetSalary(ident string) int {
	fmt.Printf("Creating fake salary for %s\n", ident)
	return rand.Intn(700_000) + 300_000
}
