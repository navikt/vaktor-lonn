package minwintid

type Stempling struct {
	StemplingTid   string `json:"Stempling_Tid"`
	Retning        string `json:"Navn"`
	Type           string `json:"Type"`
	FravarKode     int    `json:"Fravar_kode"`
	FravarKodeNavn string `json:"Fravar_kode_navn"`
}

type Stilling struct {
	Koststed  string `json:"koststed"`
	Formal    string `json:"formal"`
	Aktivitet string `json:"aktivitet"`
	RATEK001  int    `json:"RATE_K001"`
}

type Dag struct {
	Dato              string  `json:"dato"`
	SkjemaTid         float64 `json:"skjema_tid"`
	SkjemaNavn        string  `json:"skjema_navn"`
	Godkjent          int     `json:"godkjent"`
	Virkedag          string  `json:"virkedag"`
	StringStemplinger string  `json:"Stemplinger"`
	Stemplinger       []Stempling
	Stillinger        []Stilling `json:"Stillinger"`
}

type TiddataResult struct {
	VaktorNavId      string `json:"Vaktor.nav_id"`
	VaktorResourceId string `json:"Vaktor.resource_id"`
	VaktorLederNavId string `json:"Vaktor.leder_nav_id"`
	VaktorDager      string `json:"Vaktor.dager"`
	Dager            []Dag
}

type Response struct {
	VaktorVaktorTiddataResponse struct {
		VaktorVaktorTiddataResult []TiddataResult `json:"Vaktor.Vaktor_TiddataResult"`
	} `json:"Vaktor.Vaktor_TiddataResponse"`
}