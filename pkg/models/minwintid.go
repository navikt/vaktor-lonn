package models

import (
	"time"

	"github.com/shopspring/decimal"
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
	StemplingTid       string `json:"stempling_tid"`
	Retning            string `json:"navn"`
	Type               string `json:"type"`
	FravarKode         int    `json:"fravar_kode"`
	OvertidBegrunnelse string `json:"overtid_begrunnelse"`
}

type MWTStilling struct {
	Koststed      string `json:"koststed"`
	Formal        string `json:"produkt"`
	Aktivitet     string `json:"oppgave"`
	RATEK001      int    `json:"rate_k001"`
	Stillingskode string `json:"post_id"`
}

type MWTDag struct {
	Dato        string         `json:"dato"`
	SkjemaTid   float64        `json:"skjema_tid"`
	SkjemaNavn  string         `json:"skjema_navn"`
	Godkjent    int            `json:"godkjent"`
	Virkedag    string         `json:"virkedag"`
	Stemplinger []MWTStempling `json:"stemplinger"`
	Stillinger  []MWTStilling  `json:"stillinger"`
}

type MWTTiddataResult struct {
	VaktorNavId      string   `json:"nav_id"`
	VaktorResourceId string   `json:"resource_id"`
	VaktorLederNavId string   `json:"leder_nav_id"`
	VaktorLederNavn  string   `json:"leder_navn"`
	Dager            []MWTDag `json:"dager"`
}
