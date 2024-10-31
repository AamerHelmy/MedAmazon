package services

import (
	"fmt"
	"fuzzyTest/excel"
	"fuzzyTest/medName"
	"fuzzyTest/models"
	"fuzzyTest/price32"
	"sort"

	// "gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (itmSv *ItemSv) UpdateItemsFromExcelV2(filePath string) (*BulkReport, error) {
	report := &BulkReport{Name: "Base Items report"}

	// قراءة ملف Excel
	rows, err := excel.ReadFile(filePath)
	if err != nil {
		return report, fmt.Errorf("error while reading file: %v", err)
	}
	report.Total = len(rows)

	// parse excel's rows 
	excelItems, excelNames, excelSkus := itmSv.parseExcelRows(rows, report)

	// get existing data from DB
	existingSkuMap, existingNameMap := itmSv.getExistingItems(excelSkus, excelNames)

	// prepare data before save
	updateItems, newItems := itmSv.prepareUpdatesAndNewItems(excelItems, existingSkuMap, existingNameMap, report)

	// save update items
	itmSv.saveUpdates(updateItems, report)

	// save new items
	itmSv.saveNewItems(newItems, report)

	sort.Slice(report.Failed, func(i, j int) bool {
		return report.Failed[i].Row < report.Failed[j].Row
	})
	return report, nil
}

func (itmSv *ItemSv) parseExcelRows(rows [][]string, report *BulkReport) ([]ExcelItemInfo, []string, []string) {

	excelItems := make([]ExcelItemInfo, 0)
	excelNames := make([]string, 0)
	excelSkus := make([]string, 0)
	skusMap := make(map[string]int)
	nameMap := make(map[string]int)

	for i, row := range rows {
		if !itmSv.isValidRow(row, i, report) {
			continue
		}

		if itmSv.isDuplicate(row[0], row[1], i, skusMap, nameMap, report) {
			continue
		}

		item, err := itmSv.createExcelItem(row, i)
		if err != nil {
			report.Failed = append(report.Failed, ExcelItemInfo{
				Row: i + 2, SKU: row[0], Name: row[1], PricePts: 0, Error: err.Error(),
			})
			continue
		}

		skusMap[row[0]] = i
		nameMap[row[1]] = i
		excelItems = append(excelItems, item)
		excelNames = append(excelNames, row[1])
		excelSkus = append(excelSkus, row[0])
	}

	return excelItems, excelNames, excelSkus
}

func (itmSv *ItemSv) isValidRow(row []string, index int, report *BulkReport) bool {
	if len(row) == 0 {
		report.Total--
		return false
	}
	if len(row) < 3 {
		report.Failed = append(report.Failed, ExcelItemInfo{
			Row: index + 2, SKU: getSafeString(row[0], " - "), Name: getSafeString(row[1], "empty"), PricePts: 0, Error: "short data in row",
		})
		return false
	}
	if row[0] == "" || row[1] == "" || row[2] == "" {
		report.Failed = append(report.Failed, ExcelItemInfo{
			Row: index + 2, SKU: getSafeString(row[0], " - "), Name: getSafeString(row[1], "empty"), PricePts: 0, Error: "empty data in row",
		})
		return false
	}
	return true
}

func (itmSv *ItemSv) isDuplicate(sku, name string, index int, skusMap, nameMap map[string]int, report *BulkReport) bool {
	
    if prevRow, exist := skusMap[sku]; exist {
		report.Failed = append(report.Failed, ExcelItemInfo{
			Row: index + 2, SKU: sku, Name: name, PricePts: 0, Error: fmt.Sprintf("duplicate SKU found at row [%d]", prevRow+2),
		})
		return true
	}
	if prevRow, exist := nameMap[name]; exist {
		report.Failed = append(report.Failed, ExcelItemInfo{
			Row: index + 2, SKU: sku, Name: name, PricePts: 0, Error: fmt.Sprintf("duplicate Name found at row [%d]", prevRow+2),
		})
		return true
	}
	return false
}

func (itmSv *ItemSv) createExcelItem(row []string, index int) (ExcelItemInfo, error) {
	pts, err := price32.FromStringPounds(row[2])
	if err != nil {
		return ExcelItemInfo{}, fmt.Errorf("wrong price: %v", err)
	}

	return ExcelItemInfo{
		Row:       index + 2,
		SKU:       row[0],
		Name:      medName.RemoveExtraSapces(row[1]),
		CleanName: medName.Clean(row[1]),
		PricePts:  pts,
	}, nil
}

func (itmSv *ItemSv) getExistingItems(excelSkus, excelNames []string) (map[string]models.Item, map[string]models.Item) {
	existingSkuMap := make(map[string]models.Item)
	existingNameMap := make(map[string]models.Item)

	var existingSku []models.Item
	if err := itmSv.gdb.Where("sku IN ?", excelSkus).Find(&existingSku).Error; err == nil {
		for _, item := range existingSku {
			existingSkuMap[item.SKU] = item
		}
	}

	var existingNames []models.Item
	if err := itmSv.gdb.Where("name IN ?", excelNames).Find(&existingNames).Error; err == nil {
		for _, item := range existingNames {
			existingNameMap[item.Name] = item
		}
	}

	return existingSkuMap, existingNameMap
}

func (itmSv *ItemSv) prepareUpdatesAndNewItems(excelItems []ExcelItemInfo, existingSkuMap, existingNameMap map[string]models.Item, report *BulkReport) ([]models.Item, []models.Item) {
	updates := make([]models.Item, 0)
	new := make([]models.Item, 0)

	for _, excelItem := range excelItems {
		if existingSku, ok := existingSkuMap[excelItem.SKU]; ok {
			if existingSku.Name == excelItem.Name && existingSku.PricePts == excelItem.PricePts {
				report.Unchanged++
				continue
			}

			if existingName, exist := existingNameMap[excelItem.Name]; exist {
				report.Failed = append(report.Failed, ExcelItemInfo{
					Row: excelItem.Row, SKU: excelItem.SKU, Name: excelItem.Name, PricePts: excelItem.PricePts,
					Error: fmt.Sprintf("this name already exists in db at id: %d | sku: %s", existingName.ID, existingName.SKU),
				})
				continue
			}

			if existingSku.Name != excelItem.Name || existingSku.PricePts != excelItem.PricePts {
				updates = append(updates, excelItem.ToModelItem())
				report.Updated++
			}
		} else if existingName, exist := existingNameMap[excelItem.Name]; exist {
			report.Failed = append(report.Failed, ExcelItemInfo{
				Row: excelItem.Row, SKU: excelItem.SKU, Name: excelItem.Name, PricePts: excelItem.PricePts,
				Error: fmt.Sprintf("this name already exists in db at id: %d | sku: %s", existingName.ID, existingName.SKU),
			})
		} else {
			new = append(new, excelItem.ToModelItem())
			report.NewItems++
		}
	}

	return updates, new
}

func (itmSv *ItemSv) saveUpdates(updates []models.Item, report *BulkReport) {
	if len(updates) > 0 {
		updtResult := itmSv.gdb.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "sku"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "price_pts"}),
		}).CreateInBatches(updates, 1000)

		if updtResult.Error != nil {
			for _, item := range updates {
				report.Updated--
				report.Failed = append(report.Failed, ExcelItemInfo{
					SKU: item.SKU, Name: item.Name, PricePts: item.PricePts,
					Error: updtResult.Error.Error(),
				})
			}
		}
	}
}

func (itmSv *ItemSv) saveNewItems(new []models.Item, report *BulkReport) {
	if len(new) > 0 {
		err := itmSv.gdb.CreateInBatches(new, 100).Error
		if err != nil {
			for _, item := range new {
				report.NewItems--
				report.Failed = append(report.Failed, ExcelItemInfo{
					SKU: item.SKU, Name: item.Name, PricePts: item.PricePts,
					Error: err.Error(),
				})
			}
		}
	}
}
