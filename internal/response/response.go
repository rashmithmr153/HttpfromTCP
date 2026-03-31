package response

import (
	"fmt"
	"io"
	"strconv"

	"Batman/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := ""
	switch statusCode {
	case StatusOK:
		statusLine = fmt.Sprintf("HTTP/1.1 %d OK\r\n", StatusOK)
	case StatusBadRequest:
		statusLine = fmt.Sprintf("HTTP/1.1 %d Bad Request\r\n", StatusBadRequest)
	case StatusInternalServerError:
		statusLine = fmt.Sprintf("HTTP/1.1 %d Internal Server Error\r\n", StatusInternalServerError)
	default:
		return fmt.Errorf("Invalid status code")
	}
	_, err := w.Write([]byte(statusLine))
	return err
}
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	header := ""
	for k, v := range headers {
		header += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	header += "\r\n"
	_, err := w.Write([]byte(header))
	return err
}
