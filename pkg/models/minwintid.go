package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Stempling struct {
	StemplingTid string `json:"Stempling_Tid"`
	Retning      string `json:"Retning"`
	Type         string `json:"Type"`
	FravarKode   int    `json:"Fravar_kode"`
}

type Stillinger struct {
	Koststed  string `json:"koststed"`
	Formal    string `json:"formal"`
	Aktivitet string `json:"aktivitet"`
	RATEK001  int    `json:"RATE_K001"`
}

type Dag struct {
	Dato                 string       `json:"dato"`
	SkjemaTid            float64      `json:"skjema_tid"`
	SkjemaNavn           string       `json:"skjema_navn"`
	Godkjent             int          `json:"godkjent"`
	AnsattDatoGodkjentAv string       `json:"ansatt_dato_godkjent_av"`
	GodkjentDato         string       `json:"godkjent_dato"`
	Virkedag             string       `json:"virkedag"`
	Stemplinger          []Stempling  `json:"Stemplinger"`
	Stillinger           []Stillinger `json:"Stillinger"`
}

type Response struct {
	VaktorVaktorTiddataResponse struct {
		VaktorVaktorTiddataResult struct {
			VaktorRow []struct {
				VaktorNavId      string `json:"Vaktor.nav_id"`
				VaktorResourceId string `json:"Vaktor.resource_id"`
				VaktorDager      string `json:"Vaktor.dager"`
			} `json:"Vaktor.row"`
		} `json:"Vaktor.Vaktor_TiddataResult"`
	} `json:"Vaktor.Vaktor_TiddataResponse"`
}

type Clocking struct {
	In  time.Time
	Out time.Time
}

type TimeSheet struct {
	Date                time.Time
	WorkingHours        float64
	WeekendCompensation bool
	FormName            string
	Salary              decimal.Decimal
	Formal              string
	Koststed            string
	Aktivitet           string
	Clockings           []Clocking
}

type MinWinTid struct {
	Ident      string
	ResourceID string
	Timesheet  map[string]TimeSheet
	Satser     map[string]decimal.Decimal
}
