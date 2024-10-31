package services

import (
	"fmt"
	"fuzzyTest/excel"
	"fuzzyTest/models"
	"fuzzyTest/price32"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VendorSv struct {
	gdb *gorm.DB
}

func NewVenderService(db *gorm.DB) *VendorSv {
	return &VendorSv{gdb: db}
}

func (venSv *VendorSv) Create(vendor *models.Vendor) error {
	return venSv.gdb.Create(vendor).Error
}

func (venSv *VendorSv) AddNew(name string) (*models.Vendor, error) {
	vendor := models.Vendor{
		Name: name,
	}
	if err := venSv.gdb.Create(&vendor).Error; err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (venSv *VendorSv) UpdateOffersFromExcel(filePath string, vendor *models.Vendor) (*BulkReport, error) {
	report := &BulkReport{Name: vendor.Name}

	// read excel file
	rows, err := excel.ReadFile(filePath)
	if err != nil {
		return report, fmt.Errorf("error while reading file: %v", err)
	}
	report.Total = len(rows)

	offerDataList := make([]OfferData, 0, len(rows))
	skus := make([]string, 0)

	// Parse and validate all row data
	for _, row := range rows {
		if len(row) < 4 || row[0] == "" || row[1] == "" || row[2] == "" || row[3] == "" {
			continue
		}

		// parse price
		pts, err := price32.FromStringPounds(row[2])
		if err != nil {
			report.Failed = append(report.Failed, ExcelItemInfo{
				SKU: row[0], Name: row[1], PricePts: 0, Error: err.Error(),
			})
			continue
		}

		// parse discount
		discount, err := strconv.ParseFloat(row[3], 32)
		if err != nil {
			report.Failed = append(report.Failed, ExcelItemInfo{
				SKU: row[0], Name: row[1], PricePts: pts, Error: "invalid discount format",
			})
			continue
		}

		// parse quantity (optional)
		var quantity *int64
		if row[4] != "" {
			qty, err := strconv.ParseInt(row[4], 10, 64)
			if err != nil {
				report.Failed = append(report.Failed, ExcelItemInfo{
					SKU: row[0], Name: row[1], PricePts: pts, Error: "invalid quantity format",
				})
				qty = 0
			}
			quantity = &qty
		}

		offerData := OfferData{
			SKU:      row[0],
			Name:     row[1],
			PricePts: pts,
			Discount: float32(discount),
			Quantity: quantity,
		}
		offerDataList = append(offerDataList, offerData)
		skus = append(skus, row[0])
	}

	// collect existing vendor items
	existingItems := make(map[string]models.VendorItem)
	var existingVendorItems []models.VendorItem

	if err := venSv.gdb.Where("vendor_id = ? AND vendor_sku IN ?", vendor.ID, skus).Find(&existingVendorItems).Error; err != nil {
		return report, fmt.Errorf("error fetching vendor items: %v", err)
	}
	for _, item := range existingVendorItems {
		existingItems[item.VendorSKU] = item
	}

	// Prepare new vendor items and offers
	vendorItemsToSave := make([]models.VendorItem, 0)
	offersToSave := make([]models.VendorOffer, 0)

	for _, offerData := range offerDataList {
		// Check if vendor item exists
		vendorItem, exists := existingItems[offerData.SKU]

		if !exists {
			// Create new vendor item
			vendorItem = models.VendorItem{
				VendorID:   vendor.ID,
				Vendor:     *vendor,
				VendorSKU:  offerData.SKU,
				Name:       offerData.Name,
				PricePts:   offerData.PricePts,
				LastUpdate: time.Now(),
				IsActive:   true,
			}
			vendorItemsToSave = append(vendorItemsToSave, vendorItem)

			report.NewItems++

		} else if vendorItem.Name != offerData.Name || vendorItem.PricePts != offerData.PricePts {
			// collect Update existing vendor item
			vendorItem.Name = offerData.Name
			vendorItem.PricePts = offerData.PricePts
			vendorItem.LastUpdate = time.Now()
			vendorItem.IsLinked = false

			vendorItemsToSave = append(vendorItemsToSave, vendorItem)
			report.Updated++
		} else {
			report.Unchanged++
		}

		// Create offer
		offer := models.VendorOffer{
			VendorID:     vendor.ID,
			ItemSKU:      offerData.SKU,
			ItemName:     offerData.Name,
			ItemPricePts: offerData.PricePts,
			Discount:     offerData.Discount,
			Quantity:     offerData.Quantity,
			IsActive:     true,
		}
		offersToSave = append(offersToSave, offer)
	}

	if err := venSv.SaveBatchLinkedItems(vendorItemsToSave); err != nil {
		return report, err
	}

	if len(offersToSave) > 0 {
		// Mark old offers for this vendor as inactive
		if err := venSv.gdb.Model(&models.VendorOffer{}).Where("vendor_id = ?", vendor.ID).Update("is_active", false).Error; err != nil {
			return report, fmt.Errorf("error updating existing offers: %v", err)
		}

		// Create new active offers
		err := venSv.gdb.CreateInBatches(offersToSave, 1000).Error
		if err != nil {
			return report, fmt.Errorf("error creating new offers: %v", err)
		}
	}
	// match vendoritemtoSave

	return report, nil
}

// //////////////////////////////////////////////
// link single vendor item
func (venSv *VendorSv) LinkVendorItem(vendorItem *models.VendorItem) error {
	var existingLink models.VendorItem
	result := venSv.gdb.Where("vendor_id = ? AND base_item_id = ?", vendorItem.VendorID, vendorItem.BaseItemID).First(&existingLink)
	if result.Error == nil {
		return fmt.Errorf("this vendor has linked item before")
	}

	vendorItem.IsLinked = true
	vendorItem.LinkDate = time.Now()
	return venSv.gdb.Create(vendorItem).Error
}

// get all vendor's items linked in general
func (venSv *VendorSv) GetAllLinkedItems(vendorID uint) ([]models.VendorItem, error) {
	var vendorItems []models.VendorItem
	err := venSv.gdb.Where("vendor_id = ? AND is_linked = true", vendorID).
		Preload("Vendor").
		Find(&vendorItems).Error
	return vendorItems, err
}

// get base linked items by vendor id
func (venSv *VendorSv) GetBaseLinkedItemsByID(vendorID uint) ([]models.Item, error) {
	var items []models.Item

	err := venSv.gdb.
		Joins("VendorItems").Where("vendor_items.vendor_id = ? AND vendor_items.is_linked = true", vendorID).Find(&items).Error
	return items, err
}

// unlinke single verndor item
func (venSv *VendorSv) UnlinkVendorItem(vendorItemID uint) error {
	var vendorItem models.VendorItem
	err := venSv.gdb.Where("id = ?", vendorItemID).First(&vendorItem).Error
	if err != nil {
		return err
	}

	vendorItem.IsLinked = false
	vendorItem.LinkDate = time.Time{}
	return venSv.gdb.Save(&vendorItem).Error
}

// get all unlinked items by vendor id
func (venSv *VendorSv) GetAllUnlinkedItems(vendorID uint) ([]models.VendorItem, error) {
	var vendorItems []models.VendorItem
	err := venSv.gdb.Where("vendor_id = ? AND is_linked = false", vendorID).Find(&vendorItems).Error
	return vendorItems, err
}

// get all base unlinked items for vendor by vendor id
func (venSv *VendorSv) GetAllUnlinkedItemsForVendor(vendorID uint) ([]models.Item, error) {
	var items []models.Item
	err := venSv.gdb.Where("id NOT IN (?)",
		venSv.gdb.Table("vendor_items").Select("base_item_id").Where("vendor_id = ? AND is_linked = true", vendorID)).Find(&items).Error
	return items, err
}

// save batch linked items
func (venSv *VendorSv) SaveBatchLinkedItems(vendorItems []models.VendorItem) error {
	// Save vendor items first
	if len(vendorItems) > 0 {
		err := venSv.gdb.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "vendor_id"}, {Name: "vendor_sku"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "price_pts", "last_update", "is_linked"}),
		}).CreateInBatches(vendorItems, 1000).Error

		if err != nil {
			return fmt.Errorf("error saving vendor items: %v", err)
		}
	}
	return nil
}
