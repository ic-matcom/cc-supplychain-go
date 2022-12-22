package main

import (
	"time"
)

//--------------------------- Manufacture Struct -----------------------
type Manufacture struct {
	DocType          string           `json:"doctype"`
	ManufactureID    string           `json:"manufactureid"`
	OwnerID          string           `json:"ownerid"`
	AdvisorID        string           `json:"advisorid"`
	Location         Location         `json:"location"`
	Certifications   []Certification  `json:"certification"`
	Production       []string         `json:"production"`
	ToUse            []string         `json:"using"`
	ToRepair         []string         `json:"repairing"`
	ManufactureState ManufactureState `json:"manufacturestate"`
}

//------------------------- Supporting Features -------------------------------

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResultManufacture struct {
	Record    *Manufacture `json:"record"`
	TxId      string       `json:"txId"`
	Timestamp time.Time    `json:"timestamp"`
	IsDelete  bool         `json:"isDelete"`
}

type ManufactureState int16

const (
	OnProductionManufacture ManufactureState = iota
	BrokenManufacture
	DestroyedManufacture
)

//Operations
func (manufacture *Manufacture) ManufactureIsOnProduction() {
	manufacture.ManufactureState = OnProductionManufacture
}
func (manufacture *Manufacture) ManufactureIsBrokn() {
	manufacture.ManufactureState = BrokenManufacture
}
func (manufacture *Manufacture) ManufactureIsDestroyed() {
	manufacture.ManufactureState = DestroyedManufacture
}
