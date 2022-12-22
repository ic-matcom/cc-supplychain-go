package main

import (
	"time"
)

//--------------------------- Lote Struct -----------------------
type Lote struct {
	DocType           string           `json:"doctype"`
	LoteID            string           `json:"loteid"`
	ProductID         string           `json:"productid"`
	ManufactureID     string           `json:"manufactureid"`
	OwnerID           string           `json:"ownerid"`
	Advisor           string           `json:"advisor"`
	Price             Price            `json:"price"`
	Units             string           `json:"units"`
	Certifications    []Certification  `json:"certifications"`
	Environment       Environment      `json:"environment"`
	LoteState         LoteState        `json:"lotestate"`
	Components        []string         `json:"components"`
	CurrentLocationID string           `json:"currentlocationid"`
	FatherID          string           `json:"fatherid"`
	LoteProductState  LoteProductState `json:"productstate"`
}

//------------------------- Supporting Features -------------------------------

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResultLote struct {
	Record    *Lote     `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

//TraceabilityQueryResult structure usedfor returning result of traceability query
type TraceabilityQueryResultLote struct {
	TracedLoteID            string                        `json:"tracedLoteID"`
	HistoryResultLote       []HistoryQueryResultLote      `json:"historyResultLote"`
	ChildsHistoryResultLote []TraceabilityQueryResultLote `json:"chilsHistoryResultLote"`
}

type LoteState int16

const (
	ManunfacturingLote LoteState = iota
	TransportingLote
	StoredLote
	RepairingLote
	SelledLote
	DestroyedLote
	UsingLote
)

type LoteProductState int16

const (
	LoteProductNormal LoteProductState = iota
	LoteProductBroken
)

// Operations
func (lote *Lote) LoteIsManufacturing() {
	lote.LoteState = ManunfacturingLote
}
func (lote *Lote) LoteIsSelled() {
	lote.LoteState = SelledLote
}
func (lote *Lote) LoteIsTransporting() {
	lote.LoteState = TransportingLote
}
func (lote *Lote) LoteIsStored() {
	lote.LoteState = StoredLote
}
func (lote *Lote) LoteIsRepairing() {
	lote.LoteState = RepairingLote
}
func (lote *Lote) LoteIsDestroyed() {
	lote.LoteState = DestroyedLote
}
func (lote *Lote) LoteIsUsing() {
	lote.LoteState = UsingLote
}

func (lote *Lote) LoteProductIsBroken() {
	lote.LoteProductState = LoteProductBroken
}
func (lote *Lote) LoteProductIsOk() {
	lote.LoteProductState = LoteProductNormal
}
