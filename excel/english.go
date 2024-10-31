package excel

import (
    "strings"
)

// NormalizeEnglishText 
func NormalizeEnglishText(text string) string {
    text = strings.TrimSpace(text)
    text = strings.ToLower(text)
    return text
}