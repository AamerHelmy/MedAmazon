package services

import (
	"fuzzyTest/models"
	"log"
)

type MatchConfig struct {
	confidence  float32
	priceMargin float32
	batchSize   uint
	weight      Weight
}

type Weight struct {
	Leve        float32
	Partial     float32
	Token       float32
	Cosine      float32
	JaroWinkler float32
}

type MatchResult struct {
	isMatched    bool
	itemID       *uint
	itemName     string
	ItemPricePts uint32
	priceMatched bool
	confidence   float32
}

type BulkReport struct {
	Name      string
	Total     int
	NewItems  int
	Updated   int
	Unchanged int
	Failed    []ExcelItemInfo
}

type ExcelItemInfo struct {
	Row       int
	SKU       string
	Name      string
	CleanName string
	PricePts  uint32
	Discount  *float32
	Quantity  *int64
	Error     string
}

type MatchReport struct {
	Name   string
	Total  int
	Linked int
	Failed int
}

type OfferData struct {
	SKU      string
	Name     string
	PricePts uint32
	Discount float32
	Quantity *int64
}

var matchConf = MatchConfig{}

func init() {
	matchConf = MatchConfig{
		confidence:  70,
		priceMargin: 2,
		batchSize:   1000,
		weight: Weight{
			Leve:        50,
			Partial:     10,
			Token:       5,
			JaroWinkler: 35,
			Cosine:      0,
		},
	}
	if (matchConf.weight.Leve + matchConf.weight.Partial + matchConf.weight.Token + matchConf.weight.JaroWinkler + matchConf.weight.Cosine) != 100 {
		log.Fatal("err in weight not equal 100%")
	}
}

// print Bulk report
func (r *BulkReport) Print() {
	log.Printf(
		"* '%v' report:\nTotal: %d | Unchanged: %d | New: %d | Updated: %d | Failed: %d\n\n",
		r.Name, r.Total, r.Unchanged, r.NewItems, r.Updated, len(r.Failed),
	)

	if len(r.Failed) > 0 {
		log.Printf(" * Failed Items:")
		for i, item := range r.Failed {
			log.Printf("(%d)-Row[%d] -Sku[%s]: %s: %s", i+1, item.Row, item.SKU, item.Name, item.Error)
		}
	}
}

// print match report
func (r *MatchReport) Print() {
	log.Printf("# '%v' report:\nTotal: %d, Linked: %d, Failed: %d\n\n", r.Name, r.Total, r.Linked, r.Failed)
}

func getSafeString(cell string, defaultValue string) string {
	if cell != "" {
		return cell
	}
	return defaultValue
}

func getSafeInt(i *int, defaultValue int) int {
	if i != nil {
		return *i
	}
	return defaultValue
}

// ToModelItem converts ExcelItemInfo to models.Item
func (e ExcelItemInfo) ToModelItem() models.Item {
	return models.Item{
		SKU:       e.SKU,
		Name:      e.Name,
		CleanName: e.CleanName,
		PricePts:  e.PricePts,
	}
}

func (e ExcelItemInfo) ToModelVendorItem() models.VendorItem {
	return models.VendorItem{
		VendorSKU: e.SKU,
		Name:      e.Name,
		CleanName: e.CleanName,
		

		PricePts:  e.PricePts,
	}
}
