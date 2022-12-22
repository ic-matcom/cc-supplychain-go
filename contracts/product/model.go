package main

import (
	"time"
)

//--------------------------- Product Struct -----------------------
type Product struct {
	DocType            string          `json:"doctype"`
	ProductID          string          `json:"productid"`
	ProductName        string          `json:"productname"`
	Description        string          `json:"description"`
	Brand              Brand           `json:"brand"`
	Core               string          `json:"core"`
	Variety            string          `json:"variety"`
	ProductImage       string          `json:"productimage"`
	NetContent         NetContent      `json:"netcontent"`
	WidthAndHeight     Dimention       `json:"widthandheight"`
	DisplaySpace       Dimention       `json:"displayspace"`
	PackageType        string          `json:"packagetype"`
	Certifications     []Certification `json:"certifications"`
	Manufacturer       string          `json:"manufacturer"`
	ManufactureDetails string          `json:"manufacturedetails"`
	ProductState       ProductState    `json:"productstate"`
	Components         []string        `json:"components"`
	Advisor            string          `json:"advisor"`
}

//------------------------- Supporting Features -------------------------------

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResultProduct struct {
	Record    *Product  `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

type ProductState int16

const (
	OnTestingProduct ProductState = iota
	OnProductionProduct
	DiscontinuedProduct
	DestroyedProduct
)

//Operations
func (product *Product) ProductIsOnProduction() {
	product.ProductState = OnProductionProduct
}
func (product *Product) ProductIsOnTesting() {
	product.ProductState = OnTestingProduct
}
func (product *Product) ProductIsDiscontinued() {
	product.ProductState = DiscontinuedProduct
}
func (product *Product) ProductIsDesTroyed() {
	product.ProductState = DestroyedProduct
}
