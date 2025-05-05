package headers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)

	return h[key]
}

func (h *Headers) Add(key string, value string) {
	headerMap := *h
	key = strings.ToLower(key)

	oldValue := headerMap[key]
	if oldValue != "" {
		headerMap[key] = oldValue + ", " + value
	} else {
		headerMap[key] = value
	}
}

func (h *Headers) ParseHeader(bytes []byte) (bytesConsumed int, doubleCrlfFlag bool, err error) {
	str := string(bytes)

	idx1 := strings.Index(str, crlf)

	if idx1 == -1 {
		return 0, false, nil
	}

	substr := str[:idx1]

	if len(substr) == 0 {
		return len(crlf), true, nil
	}

	if !strings.Contains(substr, ":") {
		return 0, false, fmt.Errorf("':' not present in header: %s", substr)
	}

	idx2 := strings.Index(substr, ":")

	key, value := strings.ToLower(substr[:idx2]), strings.TrimSpace(substr[idx2+1:])

	validHeaderKeyRegex := regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`).MatchString

	if strings.Contains(key, " ") {
		return 0, false, fmt.Errorf("invalid spacing for key in header: %s", substr)
	}

	if !validHeaderKeyRegex(key) {
		return 0, false, fmt.Errorf("invalid character used for key in header: %s", substr)
	}

	h.Add(key, value)

	return idx1 + len(crlf), false, nil
}

func GetDefaultHeaders(contentLength int) Headers {
	h := NewHeaders()

	h.Add("content-length", strconv.Itoa(contentLength))
	h.Add("connection", "close")
	h.Add("Content-Type", "text/plain")

	return h
}
