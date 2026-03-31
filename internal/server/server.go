package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"Batman/internal/request"
	"Batman/internal/response"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (h *HandlerError) Write(w io.Writer) {
	body := []byte(h.Message)
	response.WriteStatusLine(w, h.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(body)))
	w.Write(body)

}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	request, err := request.RequestFromReader(conn)
	// fmt.Print(request)
	handleErr := &HandlerError{}
	if err != nil {
		handleErr.StatusCode = response.StatusBadRequest
		handleErr.Message = err.Error()
		handleErr.Write(conn)
		return
	}

	buff := bytes.NewBuffer([]byte{})
	handleErr = s.handler(buff, request)
	if handleErr != nil {
		handleErr.Write(conn)
		return
	}
	b := buff.Bytes()
	headers := response.GetDefaultHeaders(len(b))
	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, headers)
	conn.Write(buff.Bytes())
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("%s", err)
			return
		}
		go s.handle(conn)
	}
}

func Serve(port int, handler Handler) (*Server, error) {

	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: lister,
		handler:  handler,
	}
	go server.listen()
	return server, nil
}
