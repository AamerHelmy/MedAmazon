package services

// import (
//     "fuzzyTest/models"
//     "strings"
//     "time"
//     "unicode"

//     "github.com/paul-mannino/go-fuzzywuzzy"
//     "gorm.io/gorm/clause"
// )

// func (venSv VendorSv) AutoLinkItems(ratio float32) (*MatchReport, error) {
//     report := MatchReport{Name: "Auto Link Items"}

//     // get all items
//     var items []models.Item
//     if err := venSv.gdb.Find(&items).Error; err != nil {
//         return &report, err
//     }

//     // get all vendors's items not linked
//     var vendorItems []models.VendorItem
//     if err := venSv.gdb.Where("is_linked = ?", false).Find(&vendorItems).Error; err != nil {
//         return &report, err
//     }

//     report.Total = len(vendorItems)

//     // prepare updates
//     var toUpdate []models.VendorItem
//     for _, vi := range vendorItems {
//         MatchResult := venSv.findBestMatch(vi, items, ratio)

//         if MatchResult.isMatched {
//             vi.BaseItemID = MatchResult.itemID
//             vi.IsLinked = true
//             vi.LinkDate = time.Now()
//             vi.OurName = MatchResult.itemName
//             vi.BasePrice = MatchResult.ItemPricePts
//             report.Linked++
//         } else {
//             report.Failed++
//         }

//         toUpdate = append(toUpdate, vi)
//     }

//     // batch update
//     if len(toUpdate) > 0 {
//         if err := venSv.gdb.Clauses(clause.OnConflict{
//             Columns:   []clause.Column{{Name: "id"}},
//             DoUpdates: clause.AssignmentColumns([]string{"base_item_id", "base_price", "is_linked", "link_date", "our_name", "link_confidence"}),
//         }).CreateInBatches(toUpdate, 1000).Error; err != nil {
//             return &report, err
//         }
//     }

//     return &report, nil
// }

// func (venSv *VendorSv) findBestMatch(vi models.VendorItem, items []models.Item, ratio float32) *MatchResult {

//     viName := cleanName(vi.Name)
//     var MatchResult MatchResult

//     for _, item := range items {
//         itemName := cleanName(item.Name)

//         if isPriceMatchV1(vi.PricePts, item.PricePts) {

//             // confidence := fuzzy.Ratio(viName, itemName)
//             confidence := float32(fuzzy.Ratio(viName, itemName)) / 100
//             // confidence := float32(fuzzy.PartialRatio(viName, itemName))
//             // confidence := float32(fuzzy.TokenSortRatio(viName, itemName))

//             if confidence > MatchResult.confidence {
//                 MatchResult.confidence = confidence

//                 MatchResult.itemName = itemName
//                 MatchResult.ItemPricePts = item.PricePts
//                 MatchResult.itemID = &item.ID
//                 MatchResult.priceMatched = true
//             }
//         }
//     }

//     if MatchResult.confidence > ratio {
//         MatchResult.isMatched = true
//     }

//     return &MatchResult
// }

// // clean name
// func cleanName(text string) string {
//     var result strings.Builder
//     var lastType rune

//     for _, r := range text {
//         // no tashkiil
//         if unicode.Is(unicode.Mn, r) {
//             continue
//         }

//         // type of letter
//         currentType := 's'
//         if unicode.IsLetter(r) {
//             currentType = 'l'
//         } else if unicode.IsNumber(r) {
//             currentType = 'n'
//         }
//         // add spaces
//         switch currentType {
//         case 'l', 'n':
//             if lastType != 's' && lastType != currentType &&
//                 ((lastType == 'n' && currentType == 'l') ||
//                     (lastType == 'l' && currentType == 'n')) {
//                 result.WriteRune(' ')
//             }
//             result.WriteRune(unicode.ToLower(r))
//             lastType = currentType
//         default:
//             if lastType != 's' {
//                 result.WriteRune(' ')
//                 lastType = 's'
//             }
//         }
//     }

//     return strings.TrimSpace(result.String())
// }

// // isPriceMatch
// func isPriceMatchV1(price1, price2 uint32) bool {
//     return price1 == price2
// }