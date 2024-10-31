package models

import (
	"gorm.io/gorm"
	"time"
)

type Item struct {
	SKU       string `gorm:"unique"`
	Name      string `gorm:"unique"`
	CleanName string
	PricePts  uint32
	linked    []VendorItem
	gorm.Model
}

type Vendor struct {
	Name string `gorm:"unique"`
	gorm.Model
}

type VendorItem struct {
	VendorID   uint   `gorm:"not null;uniqueIndex:idx_vendor_code"`
	Vendor     Vendor `gorm:"foreignKey:VendorID"`
	VendorSKU  string `gorm:"not null;uniqueIndex:idx_vendor_code"`
	Name       string `gorm:"not null"`
	CleanName  string
	BaseName    string
	BaseClean   string
	WrongLink  bool
	PricePts   uint32
	BasePrice  uint32
	IsLinked   bool    `gorm:"default:false"`
	Confidence float32 `gorm:"type:decimal(6,3)"`
	BaseItemID *uint   `gorm:"index"`
	LinkDate   time.Time
	IsActive   bool `gorm:"default:true"`
	LastUpdate time.Time
	gorm.Model
}

type VendorOffer struct {
	VendorID     uint    `gorm:"not null;uniqueIndex:idx_vendor_item"`
	Vendor       Vendor  `gorm:"foreignKey:VendorID"`
	ItemSKU      string  `gorm:"not null;uniqueIndex:idx_vendor_item"`
	ItemName     string  `gorm:"not null"`
	ItemPricePts uint32  `gorm:"not null"`
	Discount     float32 `gorm:"not null;type:decimal(6,3)"`
	Quantity     *int64  `gorm:"null"`
	IsActive     bool    `gorm:"default:true"`
	gorm.Model
}
