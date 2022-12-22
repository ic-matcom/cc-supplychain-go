package main

import (
	"time"
)

//--------------------------- Warehouse Struct -----------------------
type Warehouse struct {
	DocType        string          `json:"doctype"`
	WarehouseID    string          `json:"productid"`
	OwnerID        string          `json:"ownerid"`
	AdvisorID      string          `json:"advisorid"`
	Capacity       Capacity        `json:"capacity"`
	Location       Location        `json:"location"`
	Certifications []Certification `json:"certifications"`
	Storing        []string        `json:"storing"`
	WarehouseState WarehouseState  `json:"warehousestate"`
}

//------------------------- Supporting Features -------------------------------

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResultWarehouse struct {
	Record    *Warehouse `json:"record"`
	TxId      string     `json:"txId"`
	Timestamp time.Time  `json:"timestamp"`
	IsDelete  bool       `json:"isDelete"`
}

type WarehouseState int16

const (
	WorkingWarehouse WarehouseState = iota
	NonAvailableWarehouse
	DestroyedWarehouse
)

//Operations
func (warehouse *Warehouse) WarehouseIsWorking() {
	warehouse.WarehouseState = WorkingWarehouse
}
func (warehouse *Warehouse) WarehouseIsNonAvailable() {
	warehouse.WarehouseState = NonAvailableWarehouse
}
func (warehouse *Warehouse) WarehouseIsDestroyed() {
	warehouse.WarehouseState = DestroyedWarehouse
}
