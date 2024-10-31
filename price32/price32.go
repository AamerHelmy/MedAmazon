package price32

import (
	"fmt"

	"strconv"
	"strings"
)

// StringToUint32
func FromString(s string) (uint32, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")

	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error : %v", err)
	}

	return uint32(val), nil
}

// Uint32ToString 
func ToString(n uint32) string {
	str := strconv.FormatUint(uint64(n), 10)

	return AddCommas(str)
}

// PoundsToPiasters 
func ToPiasters(pounds uint32) uint32 {
	return pounds * 100
}

// PiastersToPounds 
func ToPounds(piasters uint32) uint32 {
	return piasters / 100
}

// FormatPounds
func ToStringPounds(piasters uint32) string {
	pounds := ToPounds(piasters)
	remainingPiasters := piasters % 100

	if remainingPiasters == 0 {
		str := fmt.Sprintf("%s.00", ToString(pounds))
		return AddCommas(str)
	}

	str := fmt.Sprintf("%s.%02d", ToString(pounds), remainingPiasters)
	return AddCommas(str)
}

func FromStringPounds(s string) (uint32, error) {
	s = strings.ReplaceAll(s, ",", "")

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("wrong format")
	}

	pounds, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, err
	}

	var piasters uint64
	if len(parts) == 2 {
	
		piasterStr := parts[1]
		if len(piasterStr) == 1 {
			piasterStr += "0"
		} else if len(piasterStr) > 2 {
			piasterStr = piasterStr[:2]
		}

		piasters, err = strconv.ParseUint(piasterStr, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("error while convert pts: %v\n", err)
		}
	}

	totalPiasters := pounds*100 + piasters
	if totalPiasters > uint64(^uint32(0)) {
		return 0, fmt.Errorf("more than maximum")
	}

	return uint32(totalPiasters) , nil
}

// AddCommas
func AddCommas(str string) string {
	if len(str) < 4 {
		return str
	}

	pos := len(str) % 3
	if pos == 0 {
		pos = 3
	}

	result := str[:pos]
	for i := pos; i < len(str); i += 3 {
		result += "," + str[i:i+3]
	}

	return result
}

// RemoveNonDigits 
func RemoveNonDigits(s string) string {
	var result strings.Builder
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// IsValidUint32 
func IsValidUint32(s string) bool {
	s = RemoveNonDigits(s)
	_, err := FromString(s)
	return err == nil
}
