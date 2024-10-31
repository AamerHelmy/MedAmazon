package services

// import (
// 	"fuzzyTest/medName"
// 	"fuzzyTest/models"
// 	"time"

// 	"github.com/paul-mannino/go-fuzzywuzzy"
// 	"gorm.io/gorm/clause"
// )


// func (venSv *VendorSv) AutoLinkItemsV2() (*MatchReport, error) {
// 	report := MatchReport{Name: "Auto Link Items"}

// 	// get all items
// 	var items []models.Item
// 	if err := venSv.gdb.Find(&items).Error; err != nil {
// 		return &report, err
// 	}

// 	// get all vendors's items not linked
// 	var vendorItems []models.VendorItem
// 	if err := venSv.gdb.Where("is_linked = ?", false).Find(&vendorItems).Error; err != nil {
// 		return &report, err
// 	}

// 	report.Total = len(vendorItems)

// 	// prepare updates
// 	var toUpdate []models.VendorItem
// 	for _, vi := range vendorItems {
// 		MatchResult, err := venSv.findBestMatchV2(vi, items)
// 		if err != nil {
// 			return &report, err
// 		}

// 		vi.Confidence = MatchResult.confidence
// 		vi.OurName = MatchResult.itemName
// 		vi.BasePrice = MatchResult.ItemPricePts

// 		if MatchResult.confidence >= 60 && MatchResult.priceMatched {
// 			vi.BaseItemID = MatchResult.itemID
// 			vi.IsLinked = true
// 			vi.LinkDate = time.Now()
// 			// vi.OurName = MatchResult.itemName
// 			// vi.BasePrice = MatchResult.ItemPricePts
// 			report.Linked++
// 		} else {
// 			report.Failed++
// 		}

// 		toUpdate = append(toUpdate, vi)
// 	}

// 	// batch update
// 	if len(toUpdate) > 0 {
// 		if err := venSv.gdb.Clauses(clause.OnConflict{
// 			Columns:   []clause.Column{{Name: "id"}},
// 			DoUpdates: clause.AssignmentColumns([]string{"base_item_id", "base_price", "is_linked", "link_date", "our_name", "confidence"}),
// 		}).CreateInBatches(toUpdate, 1000).Error; err != nil {
// 			return &report, err
// 		}
// 	}

// 	return &report, nil
// }

// func (venSv *VendorSv) findBestMatchV2(vi models.VendorItem, items []models.Item) (*MatchResult, error) {

// 	viName := medName.Clean(vi.Name)
// 	var MatchResult MatchResult

// 	for _, item := range items {
// 		itemName := medName.Clean(item.Name)
// 		if isPriceMatchV2(vi.PricePts, item.PricePts) {

// 			ratioScore := float32(fuzzy.Ratio(viName, itemName))
// 			partialScore := float32(fuzzy.PartialRatio(viName, itemName))
// 			tokenSortScore := float32(fuzzy.TokenSortRatio(viName, itemName))

// 			confidence := (ratioScore*0.55 + partialScore*0.25 + tokenSortScore*0.20)

// 			// update best match
// 			if confidence > MatchResult.confidence {
// 				MatchResult.itemID = &item.ID
// 				MatchResult.itemName = itemName
// 				MatchResult.ItemPricePts = item.PricePts
// 				MatchResult.confidence = confidence
// 				MatchResult.priceMatched = true
// 			}
// 		}
// 	}

// 	return &MatchResult, nil
// }

// // isPriceMatch
// func isPriceMatchV2(price1, price2 uint32) bool {
// 	return price1 == price2
// }
