package dummy

import (
	"fmt"
	"math/rand"
)

func generateWorkhours() string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", rand.Intn(4)+6, rand.Intn(60), rand.Intn(4)+14, rand.Intn(60))
}

func GetMinWinTid(ident string) map[string][]string {
	fmt.Printf("Creating fake timesheet for %s\n", ident)
	return map[string][]string{
		"14.03.2022": {generateWorkhours()},
		"15.03.2022": {generateWorkhours()},
		"16.03.2022": {"01:00-03:00", generateWorkhours()},
		"17.03.2022": {generateWorkhours()},
		"18.03.2022": {generateWorkhours()},
		"19.03.2022": {},
		"20.03.2022": {},
	}
}

func GetSalary(ident string) int {
	fmt.Printf("Creating fake salary for %s\n", ident)
	return rand.Intn(700_000) + 300_000
}
