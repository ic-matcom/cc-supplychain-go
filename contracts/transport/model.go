package main

import (
	"time"
)

//--------------------------- Transport Struct -----------------------
type Transport struct {
	DocType        string          `json:"doctype"`
	TransportID    string          `json:"transportid"`
	OwnerID        string          `json:"ownerid"`
	AdvisorID      string          `json:"advisorid"`
	Capacity       Capacity        `json:"capacity"`
	Location       Location        `json:"location"`
	TransportState TransportState  `json:"transportstate"`
	TransportType  string          `json:"transporttype"`
	Certifications []Certification `json:"certifications"`
	Transporting   Transporting    `json:"transporting"`
	Shipment       []string        `json:"shipment"`
}

//------------------------- Supporting Features -------------------------------

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResultTransport struct {
	Record    *Transport `json:"record"`
	TxId      string     `json:"txId"`
	Timestamp time.Time  `json:"timestamp"`
	IsDelete  bool       `json:"isDelete"`
}

type TransportState int16

const (
	AvailableTransport TransportState = iota
	LoadingTrasport
	DeliveringTransport
	NonAvailableTransport
	DestroyedTransport
)

//Operations
func (transport *Transport) TransportIsAvailable() {
	transport.TransportState = AvailableTransport
}
func (transport *Transport) TransportIsDelivering() {
	transport.TransportState = DeliveringTransport
}
func (transport *Transport) TransportIsLoading() {
	transport.TransportState = LoadingTrasport
}
func (transport *Transport) TransportIsNonAvailable() {
	transport.TransportState = NonAvailableTransport
}

func (transport *Transport) TransportIsDestroyed() {
	transport.TransportState = DestroyedTransport
}
