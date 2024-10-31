package excel

import (
    "strings"
)

// NormalizeArabicText 
func NormalizeArabicText(text string) string {
    text = removeDiacritics(text)
    text = normalizeArabicLetters(text)
    return text
}

// removeDiacritics 
func removeDiacritics(text string) string {
    return strings.Map(func(r rune) rune {
        if r >= 0x064B && r <= 0x065F {
            return -1
        }
        return r
    }, text)
}

// normalizeArabicLetter
func normalizeArabicLetters(text string) string {
    replacer := strings.NewReplacer(
        "ﺃ", "أ",
        "ﺁ", "آ",
        "ﺇ", "إ",
        "ﺍ", "ا",
        "ﻯ", "ى",
        "ﺉ", "ئ",
        "ﺓ", "ة",
    )
    return replacer.Replace(text)
}