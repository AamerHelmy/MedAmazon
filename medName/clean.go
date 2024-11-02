package medName

import (
	"strconv"
	"strings"
	"unicode"
)

// Main Clean : كيتوفاااااان سعر قدديييييييم
// - Remove Extra Spaces
// - Removes all diacritical marks from the text
// - Remove Unwanted Words
// - Converting to lowercase
func Clean(text string) string {
	if text == "" {
		return ""
	}
	return RemoveUnwantedWords(RemoveExtraLettersArLatin(ExtractArabicLatin(text)))
}

// Keep Arabc and Latin letters only
func ExtractArabicLatin(text string) string {
	var result strings.Builder
	result.Grow(len(text))

	for _, r := range text {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.Is(arabicRange, r) || unicode.In(r, unicode.Latin) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return result.String()
}

// Remove Unwanted Words
func RemoveUnwantedWords(text string) string {
	words := strings.Fields(text)
	result := make([]string, 0, len(words))

	for _, word := range words {
		if _, exists := unwantedWords[strings.ToLower(word)]; !exists {
			result = append(result, word)
		}
	}
	return strings.Join(result, " ")
}

func RemoveExtraLettersArLatin(text string) string {
	if len(text) <= 1 {
		return text
	}
	var result strings.Builder
	result.Grow(len(text))
	var previous rune

	for _, current := range text {
		if current != previous {
			result.WriteRune(current)
			previous = current
		}
	}
	return result.String()
}

func RemoveExtraLettersLatin(text string) string {
	const bufferSize = 64
	var result strings.Builder
	result.Grow(len(text))

	buffer := make([]byte, 0, bufferSize)
	count := 1
	buffer = append(buffer, text[0])

	for i := 1; i < len(text); i++ {
		if text[i] == text[i-1] {
			count++
		} else {
			count = 1
		}
		if count < 2 {
			buffer = append(buffer, text[i])
		}
		if len(buffer) >= bufferSize {
			result.Write(buffer)
			buffer = buffer[:0]
		}
	}
	if len(buffer) > 0 {
		result.Write(buffer)
	}
	return result.String()
}

func RemoveExtraSapces(text string) string {
	words := strings.Fields(text)
	var result strings.Builder
	result.Grow(len(text))
	for _, word := range words {
		result.WriteString(word)
		result.WriteRune(' ')
	}
	return result.String()
}

// Adding spaces between numbers and letters
func SeparateNameAndNumber(text string) string {
	words := strings.Fields(text)
	var result strings.Builder
	result.Grow(len(text))

	for i, word := range words {
		if i > 0 {
			result.WriteRune(' ')
		}
		formatWord(&result, word)
	}
	return result.String()
}

// extract concertaon as string
func ExtractConcentration(name string) string {
	name = SeparateNameAndNumber(name)
	var result strings.Builder

	parts := strings.Fields(name)
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			result.WriteString(part)
			result.WriteRune(' ')
		}
	}
	return result.String()
}

// formatWord formats a single word by adding spaces between
func formatWord(result *strings.Builder, word string) {
	var lastType charType
	for _, r := range word {
		currentType := getCharType(r)

		// Add space between numbers and letters
		if addSpace(lastType, currentType) {
			result.WriteRune(' ')
		}

		result.WriteRune(unicode.ToLower(r))
		lastType = currentType
	}
}

// getCharType determines the type of the given rune
func getCharType(r rune) charType {
	switch {
	case unicode.IsLetter(r):
		return letterType
	case unicode.IsNumber(r):
		return numberType
	default:
		return symbolType
	}
}

// AddSpace determines if a space should be added
func addSpace(lastType, currentType charType) bool {
	return lastType != "" &&
		lastType != currentType &&
		((lastType == numberType && currentType == letterType) ||
			(lastType == letterType && currentType == numberType))
}
