package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string 

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	clrfIdx := strings.Index(string(data), "\r\n")
	if clrfIdx == -1 {
		return 0, false, nil
	}

	if clrfIdx == 0 {
		return 2, true, nil
	}

	fieldLine := string(data[:clrfIdx])
	fieldLine = strings.TrimSpace(fieldLine)
	if len(fieldLine) == 0 {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	semiColonIdx := strings.Index(fieldLine, ":")
	if semiColonIdx == -1 {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	key := fieldLine[:semiColonIdx]
	key, correct := isHeaderKeyCorrect(key)
	if !correct {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	val := fieldLine[semiColonIdx+1:]
	val, correct = isHeaderValCorrect(val)
	if !correct {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	h[key] = val
	return clrfIdx+2, false, nil
}

func isHeaderKeyCorrect(key string) (string, bool) {
	if len(key) == 0 {
		return key, false
	}

	for _, ch := range key {
		if ch == ' ' {
			return key, false
		}
	}

	return key, true 
}

func isHeaderValCorrect(val string) (string, bool) {
	if len(val) == 0 {
		return val, false
	}

	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return val, false
	}

	return val, true
}