package response

import (
	"fmt"
	"io"
	"tcp2http/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func status(code StatusCode) string {
	switch code {
	case StatusOK:
		return "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		return "HTTP/1.1 400 Bad Request\r\n"
	default:
		return "HTTP/1.1 500 Internal Server Error\r\n"
	}
}

type Writer struct {
	w io.Writer
}

func New(w io.Writer) *Writer {
	return &Writer{w: w}
}

func DefaultHeaders(len int, ct string, keepAlive bool) *headers.Headers {
	h := headers.New()
	h.Set("Content-Length", fmt.Sprintf("%d", len))
	h.Set("Content-Type", ct)
	if keepAlive {
		h.Set("Connection", "keep-alive")
	} else {
		h.Set("Connection", "close")
	}
	return h
}

func (wr *Writer) Write(code StatusCode, h *headers.Headers, body []byte) {
	wr.w.Write([]byte(status(code)))
	wr.w.Write(headers.Format(h))
	if body != nil {
		wr.w.Write(body)
	}
}

func (wr *Writer) WriteChunk(b []byte) {
	fmt.Fprintf(wr.w, "%x\r\n", len(b))
	wr.w.Write(b)
	wr.w.Write([]byte("\r\n"))
}

func (wr *Writer) EndChunked() {
	wr.w.Write([]byte("0\r\n\r\n"))
}

func (wr *Writer) WriteRaw(b []byte) {
	wr.w.Write(b)
}
