package models

type Period struct {
	Fra       string `json:"fra"`
	Til       string `json:"til"`
	Helligdag bool   `json:"helligdag"`
}

type Plan struct {
	Ident   string            `json:"ident"`
	Satser  map[string]int    `json:"satser"`
	Periods map[string]Period `json:"periods"`
}
