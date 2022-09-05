package dummy

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"math/rand"
)

func generateWorkhours() string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", rand.Intn(4)+6, rand.Intn(60), rand.Intn(4)+14, rand.Intn(60))
}

func GetMinWinTid(ident string) map[string][]string {
	log.Printf("Creating fake timesheet for %s", ident)
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
