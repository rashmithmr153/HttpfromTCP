package request

import (
	"Batman/internal/headers"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ParserState int

const (
	stateInitialized ParserState = iota
	stateParseHeader
	stateParseBody
	stateDone
)

type Request struct {
	RequestLine RequestLine
	Header      headers.Headers
	Body        []byte
	State       ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var LINE_SEP = "\r\n"
var BAD_REQ = fmt.Errorf("Bad request line")

const BUFF_SIZE = 8

func getInt(h headers.Headers, name string, defaultValue int) int {
	valueStr, exists := h.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	for {
		switch r.State {
		case stateInitialized:
			rl, lenRead, err := parseRequestLine(string(data))

			if err != nil {
				return 0, err
			}

			if lenRead == 0 {
				return 0, nil
			}

			r.RequestLine = *rl
			read += lenRead
			r.State = stateParseHeader

		case stateParseHeader:
			n, done, err := r.Header.Parse(data[read:])
			if err != nil {
				return 0, err
			}
			read += n
			if !done {
				return read, nil
			}
			r.State = stateParseBody
		case stateParseBody:
			length := getInt(r.Header, "content-length", 0)
			if length == 0 {
				r.State = stateDone
				return read, nil
			}

			remaining := min(length-len(r.Body), len(data[read:]))
			r.Body = append(r.Body, data[read:read+remaining]...)
			read += remaining
			if len(r.Body) == length {
				r.State = stateDone
				return read, nil
			}

		case stateDone:
			return 0, fmt.Errorf("error: trying to read data in a done state")
		default:
			return 0, fmt.Errorf("error:state")
		}
	}
}

func parseRequestLine(request string) (*RequestLine, int, error) {
	idx := strings.Index(request, LINE_SEP)
	if idx == -1 {
		return nil, 0, nil
	}
	var partsOfRequest RequestLine
	//POST /coffee HTTP/1.1
	startLine := request[:idx]

	bytesRead := idx + len(LINE_SEP)
	parts := strings.Split(startLine, " ")

	if len(parts) != 3 {
		return nil, bytesRead, BAD_REQ
	}

	httpParts := strings.Split(parts[2], "/")

	if len(httpParts) != 2 || httpParts[1] != "1.1" || httpParts[0] != "HTTP" {
		return nil, bytesRead, BAD_REQ
	}
	partsOfRequest.Method = parts[0]
	partsOfRequest.RequestTarget = parts[1]
	partsOfRequest.HttpVersion = httpParts[1]

	return &partsOfRequest, bytesRead, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{
		State:  stateInitialized,
		Header: headers.NewHeaders(),
	}
	buff := make([]byte, BUFF_SIZE)
	var readIndex = 0
	for {
		if request.State == stateDone {
			break
		}

		if readIndex == len(buff) {
			newBuff := make([]byte, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}
		readLen, err := reader.Read(buff[readIndex:])
		if err != nil {
			if err == io.EOF {
				if request.State == stateParseBody {
					length := getInt(request.Header, "content-length", 0)
					if len(request.Body) != length {
						return nil, fmt.Errorf("incomplete request body")
					}
				}
				request.State = stateDone
				break
			}
			return nil, errors.Join(
				fmt.Errorf("Error in while data reading into buffer:"), err)
		}
		readIndex += readLen
		parsLen, err := request.parse(buff[:readIndex])

		if err != nil {
			return nil, err
		}
		copy(buff, buff[parsLen:readIndex])
		readIndex -= parsLen
	}
	return &request, nil
}
