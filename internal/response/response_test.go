package response

import (
	"bytes"
	"pier/internal/headers"
	"testing"
)

func TestWriteBasicResponse(t *testing.T) {
	buf := &bytes.Buffer{}

	w := New(buf)

	body := []byte("hello")
	h := DefaultHeaders(len(body), "text/plain", false)

	w.Write(StatusOK, h, body)

	out := buf.String()

	if !contains(out, "HTTP/1.1 200 OK\r\n") {
		t.Fatalf("missing status")
	}
	if !contains(out, "Content-Length: 5") {
		t.Fatalf("missing length")
	}
	if !contains(out, "hello") {
		t.Fatalf("missing body")
	}
}

func TestChunkedWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	w := New(buf)

	w.WriteChunk([]byte("abc"))
	w.EndChunked()

	out := buf.String()

	if !contains(out, "3\r\nabc\r\n") {
		t.Fatalf("invalid chunk format")
	}
	if !contains(out, "0\r\n\r\n") {
		t.Fatalf("missing chunk terminator")
	}
}

func TestHeaderFormatting(t *testing.T) {
	h := headers.New()
	h.Set("Content-Type", "text/plain")

	buf := &bytes.Buffer{}
	buf.Write(headers.Format(h))

	if !contains(buf.String(), "Content-Type: text/plain") {
		t.Fatalf("wrong header output")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) &&
		(len(s) == len(sub) && s == sub || (len(s) > len(sub) && (s[:len(sub)] == sub || contains(s[1:], sub))))
}
