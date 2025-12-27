package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"tcp2http/internal/headers"
	"time"
)

type Request struct {
	Method  string
	Target  string
	Version string
	Headers *headers.Headers
	Body    []byte
}

var (
	ErrMalformed = errors.New("malformed request")
	ErrTimeout   = errors.New("request timeout")
)

func readUntil(r io.Reader, delim []byte, timeout time.Duration) ([]byte, error) {
	buf := make([]byte, 0, 2048)
	tmp := make([]byte, 512)

	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return nil, ErrTimeout
		}

		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if bytes.Contains(buf, delim) {
				return buf, nil
			}
		}

		if err != nil {
			if err == io.EOF {
				return buf, nil
			}
			return nil, err
		}
	}
}

func parseRequestLine(line string) (string, string, string, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "", "", "", ErrMalformed
	}
	if !strings.HasPrefix(parts[2], "HTTP/") {
		return "", "", "", ErrMalformed
	}
	return parts[0], parts[1], strings.TrimPrefix(parts[2], "HTTP/"), nil
}

func parseChunked(r io.Reader) ([]byte, error) {
	var body []byte
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
			if bytes.Contains(body, []byte("\r\n0\r\n")) {
				return body, nil
			}
		}
		if err != nil {
			return body, err
		}
	}
}

func RequestFromReader(r io.Reader) (*Request, error) {
	data, err := readUntil(r, []byte("\r\n\r\n"), 5*time.Second)
	if err != nil {
		return nil, err
	}

	idx := bytes.Index(data, []byte("\r\n"))
	if idx < 0 {
		return nil, ErrMalformed
	}

	line := strings.TrimRight(string(data[:idx]), "\r\n")
	method, target, version, err := parseRequestLine(line)
	if err != nil {
		return nil, err
	}

	h, n, err := headers.Parse(data[idx+2:])
	if err != nil {
		return nil, err
	}

	bodyStart := idx + 2 + n
	body := data[bodyStart:]

	req := &Request{Method: method, Target: target, Version: version, Headers: h}

	if te, ok := h.Get("Transfer-Encoding"); ok && strings.Contains(strings.ToLower(te), "chunked") {
		chunked, _ := parseChunked(r)
		req.Body = append(body, chunked...)
		return req, nil
	}

	if cl, ok := h.Get("Content-Length"); ok {
		n, _ := strconv.Atoi(cl)
		remaining := n - len(body)
		if remaining > 0 {
			buf := make([]byte, remaining)
			io.ReadFull(r, buf)
			body = append(body, buf...)
		}
	}

	req.Body = body
	return req, nil
}
