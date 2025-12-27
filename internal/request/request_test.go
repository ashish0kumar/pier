package request

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestParseBasicRequest(t *testing.T) {
	raw := "" +
		"GET /test HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"\r\n"

	r, err := RequestFromReader(bytes.NewBufferString(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Method != "GET" {
		t.Fatalf("wrong method")
	}
	if r.Target != "/test" {
		t.Fatalf("wrong target")
	}
	if r.Version != "1.1" {
		t.Fatalf("wrong http version")
	}
}

func TestMalformedRequestLine(t *testing.T) {
	raw := "BROKEN\r\n\r\n"

	_, err := RequestFromReader(bytes.NewBufferString(raw))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRequestWithBody(t *testing.T) {
	raw := "" +
		"POST /data HTTP/1.1\r\n" +
		"Host: x\r\n" +
		"Content-Length: 5\r\n" +
		"\r\n" +
		"hello"

	r, err := RequestFromReader(bytes.NewBufferString(raw))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	if string(r.Body) != "hello" {
		t.Fatalf("expected body hello, got %q", string(r.Body))
	}
}

func TestChunkedRequestBody(t *testing.T) {
	raw := "" +
		"POST /c HTTP/1.1\r\n" +
		"Host: x\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\n" +
		"5\r\nhello\r\n" +
		"0\r\n\r\n"

	r, err := RequestFromReader(bytes.NewBufferString(raw))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	if !strings.Contains(string(r.Body), "hello") {
		t.Fatalf("expected chunked content")
	}
}

type slowReader struct{}

func (s slowReader) Read(b []byte) (int, error) {
	time.Sleep(6 * time.Second)
	return 0, io.EOF
}

func TestTimeout(t *testing.T) {
	_, err := RequestFromReader(slowReader{})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}
