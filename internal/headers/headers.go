package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

var CRLF = []byte("\r\n")

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(name string) string {
	return h[strings.ToLower(name)]
}

func (h Headers) Set(name, value string) {
	h[strings.ToLower(name)] = value
}

func isToken(str string) bool {
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' {
			continue
		}
		// a-z
		if ch >= 'a' && ch <= 'z' {
			continue
		}
		// 0-9
		if ch >= '0' && ch <= '9' {
			continue
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	//"Host: localhost:42069\r\n\r\n"
	headerParts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(headerParts) != 2 {
		return "", "", fmt.Errorf("Not valid header-line")
	}

	name := headerParts[0]
	value := bytes.TrimSpace(headerParts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("Not valid header-name")
	}

	return string(name), string(value), nil

}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data, CRLF)
		if idx == -1 {
			break
		}
		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}
		name, value, err := parseHeader(data[:idx])
		if err != nil {
			return read, false, err
		}
		if !isToken(name) {
			return read, false, fmt.Errorf("Invalid header name")
		}
		read += idx + len(CRLF)
		data = data[read:]
		h.Set(name, value)
	}
	return read, done, nil
}
