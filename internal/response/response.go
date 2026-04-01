package response

import (
	"fmt"
	"io"
	"strconv"

	"Batman/internal/headers"
)

type StatusCode int
type writerState int

type Writer struct {
	writer io.Writer
	state  writerState
}

const (
	stateStatusLine writerState = iota
	stateHeaders
	stateBody
	stateDone
)

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  stateStatusLine,
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {

	if w.state != stateStatusLine {
		return fmt.Errorf("can not write status line in current state")
	}

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
	_, err := w.writer.Write([]byte(statusLine))
	w.state = stateHeaders
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {

	if w.state != stateHeaders {
		return fmt.Errorf("can not write headers in current state")
	}

	header := ""
	for k, v := range headers {
		header += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	header += "\r\n"
	_, err := w.writer.Write([]byte(header))
	w.state = stateBody
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("can not write body in current state")
	}
	n, err := w.writer.Write(p)
	w.state = stateDone
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("can not write body in current state")
	}
	size := len(p)
	hexSize := fmt.Sprintf("%x\r\n", size)
	n, err := w.writer.Write([]byte(hexSize + string(p) + "\r\n"))
	return n, err
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != stateBody {
		return 0, fmt.Errorf("can not write body in current state")
	}
	n, err := w.writer.Write([]byte("0\r\n\r\n"))
	w.state = stateDone
	return n, err
}
