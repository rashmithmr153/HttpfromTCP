package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var CRLF = []byte("\r\n")

func NewHeaders() Headers {
	return Headers{}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	//"Host: localhost:42069\r\n\r\n"
	headerParts := bytes.SplitN(fieldLine, []byte(":"), 2)
	fmt.Println("_________lenght: ", len(headerParts))
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
		read += idx + len(CRLF)
		data = data[read:]
		h[name] = value
	}
	return read, done, nil
}
