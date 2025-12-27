package server

import (
	"fmt"
	"net"
	"pier/internal/request"
	"pier/internal/response"
	"time"
)

type Handler func(w *response.Writer, r *request.Request)
type Middleware func(next Handler) Handler

type Server struct {
	l net.Listener
	m []Middleware
	h Handler
}

func wrap(h Handler, m []Middleware) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func Serve(port uint16, h Handler, middleware ...Middleware) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{l: ln, m: middleware, h: wrap(h, middleware)}
	go s.loop()
	return s, nil
}

func (s *Server) loop() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()

	for {
		_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
		r, err := request.RequestFromReader(c)

		if err != nil {
			writer := response.New(c)
			headers := response.DefaultHeaders(0, "text/plain", false)
			body := []byte("Bad Request")
			writer.Write(response.StatusBadRequest, headers, body)
			return
		}

		w := response.New(c)
		s.h(w, r)

		if conn, ok := r.Headers.Get("Connection"); ok && conn == "close" {
			return
		}
	}
}

func (s *Server) Close() error {
	return s.l.Close()
}
