package services

import (
	"fmt"
	"fuzzyTest/medName"
	"fuzzyTest/models"
	"time"

	"github.com/paul-mannino/go-fuzzywuzzy"
	"github.com/xrash/smetrics"
	"gorm.io/gorm/clause"
)

func (venSv *VendorSv) AutoLinkItemsV3() (*MatchReport, error) {
	report := MatchReport{
		Name: "Auto Link Items",
	}

	// get all items
	var items []models.Item
	if err := venSv.gdb.Find(&items).Error; err != nil {
		return &report, fmt.Errorf("failed to fetch items: %w", err)
	}

	// get all vendors's items not linked
	var vendorItems []models.VendorItem
	if err := venSv.gdb.Where("is_linked = ?", false).Find(&vendorItems).Error; err != nil {
		return &report, fmt.Errorf("failed to fetch vendor items: %w", err)
	}

	report.Total = len(vendorItems)
	toUpdate := make([]models.VendorItem, 0, len(vendorItems))

	for _, vi := range vendorItems {
		MatchResult := venSv.findBestMatchV3(vi, items)

		vi.Confidence = MatchResult.confidence
		vi.BaseName = MatchResult.itemName
		vi.BasePrice = MatchResult.ItemPricePts

		if MatchResult.confidence >= matchConf.confidence && MatchResult.priceMatched {
			vi.BaseItemID = MatchResult.itemID
			vi.IsLinked = true
			vi.LinkDate = time.Now()
			report.Linked++
		} else {
			report.Failed++
		}

		toUpdate = append(toUpdate, vi)
	}

	if len(toUpdate) > 0 {
		if err := venSv.gdb.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"base_item_id", "base_price", "is_linked", "link_date", "our_name", "confidence",
			}),
		}).CreateInBatches(toUpdate, int(matchConf.batchSize)).Error; err != nil {
			return &report, fmt.Errorf("failed to update items: %w", err)
		}
	}
	return &report, nil
}

func (venSv *VendorSv) findBestMatchV3(vi models.VendorItem, items []models.Item) *MatchResult {
	viName := medName.Clean(vi.Name)
	var MatchResult MatchResult

	for _, item := range items {
		if IsPriceMatchV3(vi.PricePts, item.PricePts) {
			itemName := medName.Clean(item.Name)
			confidence := calculateConfidence(viName, itemName)

			if confidence > MatchResult.confidence {
				itemID := item.ID
				MatchResult.itemID = &itemID
				MatchResult.itemName = itemName
				MatchResult.ItemPricePts = item.PricePts
				MatchResult.confidence = confidence
				MatchResult.priceMatched = true
			}
		}
	}

	return &MatchResult
}

func calculateConfidence(str1, str2 string) float32 {
	leveScore := float32(fuzzy.Ratio(str1, str2)) / 100
	partialScore := float32(fuzzy.PartialRatio(str1, str2)) / 100
	tokenScore := float32(fuzzy.TokenSortRatio(str1, str2)) / 100
	jaroScore := float32(smetrics.JaroWinkler(str1, str2, 0.7, 4)*100) / 100
	// cosineScore := strutil.Similarity("think", "tank", metrics.)

	// fmt.Printf("leveScore: %f ,PartialScore: %f tokenSortScore: %f jaroWinkler: %f\n", leveScore, partialScore, tokenScore, jaroScore)

	return (leveScore * (matchConf.weight.Leve)) +
		(partialScore * (matchConf.weight.Partial)) +
		(tokenScore * (matchConf.weight.Token)) +
		(jaroScore * matchConf.weight.JaroWinkler)
}

// isPriceMatch
func IsPriceMatchV3(price1, price2 uint32) bool {
	dif := float32(price1) - float32(price2)
	if dif < 0 {
		dif = float32(price2) - float32(price1)
	} else if dif == 0 {
		return true
	}

	if ((dif / float32(price1)) * 100) < matchConf.priceMargin {
		return true
	}
	return false
}
