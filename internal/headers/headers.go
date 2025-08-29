package headers

import (
	"fmt"
	"strings"
)

type HeaderKeySet map[rune]bool
type Headers map[string]string 

func (hks HeaderKeySet) initialize() {
	for ch := 'a'; ch <= 'z'; ch++ {
		hks[ch] = true
	}

	for ch := 'A'; ch <= 'Z'; ch++ {
		hks[ch] = true
	}

	for ch := '0'; ch <= '9'; ch++ {
		hks[ch] = true
	}

	hks['!'] = true;
	hks['#'] = true;
	hks['$'] = true;
	hks['%'] = true;
	hks['&'] = true;
	hks['\''] = true;
	hks['*'] = true;
	hks['+'] = true;
	hks['-'] = true;
	hks['.'] = true;
	hks['^'] = true;
	hks['_'] = true;
	hks['|'] = true;
	hks['~'] = true;
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	hks := make(HeaderKeySet)
    hks.initialize() 

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
	key, correct := isHeaderKeyCorrect(key, hks)
	if !correct {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	val := fieldLine[semiColonIdx+1:]
	val, correct = isHeaderValCorrect(val)
	if !correct {
		return 0, false, fmt.Errorf("header field line is incorrect")
	}

	_val, ok := h[key] 
	if !ok {
		h[key] = val
	} else {
		h[key] = _val + ", " + val
	}
	
	clrfIdx2 := strings.Index(string(data), "\r\n\r\n")
	if clrfIdx2 == clrfIdx {
		return clrfIdx + 4, true, nil 
	}
	return clrfIdx+2, false, nil
}

func isHeaderKeyCorrect(key string, hks HeaderKeySet) (string, bool) {
	if len(key) == 0 {
		return key, false
	}

	for _, ch := range key {
		if !hks[ch] {
			return key, false
		}
	}

	key = strings.ToLower(key)
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