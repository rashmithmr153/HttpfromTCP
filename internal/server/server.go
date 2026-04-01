package server

import (
	"fmt"
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

type Handler func(w *response.Writer, req *request.Request)

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	request, err := request.RequestFromReader(conn)
	writer := response.NewWriter(conn)
	if err != nil {
		body := []byte(err.Error())
		headers := response.GetDefaultHeaders(len(body))

		writer.WriteStatusLine(response.StatusBadRequest)
		writer.WriteHeaders(headers)
		writer.WriteBody(body)
		return
	}
	s.handler(writer, request)
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
