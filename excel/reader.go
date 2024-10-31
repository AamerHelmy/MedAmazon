package excel

import (
	"fmt"
	"unicode"

	"github.com/xuri/excelize/v2"
)

// Read Excel file
func ReadFile(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("not found any sheets in excel file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("error while reading rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	var validRows [][]string
	for i, row := range rows {
		if i == 0 {
			continue
		}
		var rowData []string
		for j := 0; j < len(row); j++ {
			text := row[j]
			text = NormalizeText(text)
			rowData = append(rowData, text)
		}
		validRows = append(validRows, rowData)
	}

	return validRows, nil
}

// NormalizeText
func NormalizeText(text string) string {
	if isArabic(text) {
		return NormalizeArabicText(text)
	}
	return NormalizeEnglishText(text)
}

func isArabic(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Arabic) {
			return true
		}
	}
	return false
}

// func isNumeric(s string) bool {
// 	for _, r := range s {
// 		if !unicode.IsDigit(r) {
// 			return false
// 		}
// 	}
// 	return true
// }
