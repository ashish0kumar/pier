package request

import (
	"bytes"
	"fmt"
	"io"
	"tcp2http/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var crlf = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, crlf)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(crlf)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		currentData := data[read:]

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, nil
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.state = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("error")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: buffer could get overrun by a header that exceeds 1k or the body
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
