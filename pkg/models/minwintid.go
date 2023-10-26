package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Clocking struct {
	In  time.Time
	Out time.Time
	OtG bool
}

type TimeSheet struct {
	Date          time.Time
	WorkingHours  float64
	WorkingDay    string
	FormName      string
	Salary        decimal.Decimal
	Stillingskode string
	Formal        string
	Koststed      string
	Aktivitet     string
	Clockings     []Clocking
}

type MinWinTid struct {
	Ident        string
	ResourceID   string
	ApproverID   string
	ApproverName string
	Timesheet    map[string]TimeSheet
	Satser       Satser
}

type MWTStempling struct {
	StemplingTid       string `json:"Stempling_Tid"`
	Retning            string `json:"Navn"`
	Type               string `json:"Type"`
	FravarKode         int    `json:"Fravar_kode"`
	OvertidBegrunnelse string `json:"Overtid_Begrunnelse"`
}

type MWTStilling struct {
	Koststed      string `json:"koststed"`
	Formal        string `json:"formal"`
	Aktivitet     string `json:"aktivitet"`
	RATEK001      int    `json:"RATE_K001"`
	Stillingskode string `json:"post_id"`
}

type MWTDag struct {
	Dato        string         `json:"dato"`
	SkjemaTid   float64        `json:"skjema_tid"`
	SkjemaNavn  string         `json:"skjema_navn"`
	Godkjent    int            `json:"godkjent"`
	Virkedag    string         `json:"virkedag"`
	Stemplinger []MWTStempling `json:"Stemplinger"`
	Stillinger  []MWTStilling  `json:"Stillinger"`
}

type MWTTiddataResult struct {
	VaktorNavId      string `json:"Vaktor.nav_id"`
	VaktorResourceId string `json:"Vaktor.resource_id"`
	VaktorLederNavId string `json:"Vaktor.leder_nav_id"`
	VaktorLederNavn  string `json:"Vaktor.leder_navn"`
	VaktorDager      string `json:"Vaktor.dager"`
	Dager            []MWTDag
}

type MWTResponse struct {
	VaktorVaktorTiddataResponse struct {
		VaktorVaktorTiddataResult []MWTTiddataResult `json:"Vaktor.Vaktor_TiddataResult"`
	} `json:"Vaktor.Vaktor_TiddataResponse"`
}
