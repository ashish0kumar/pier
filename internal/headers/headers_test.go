package headers

import (
	"testing"
)

func TestParseSimpleHeaders(t *testing.T) {
	raw := []byte("Host: example.com\r\nUser-Agent: test\r\n\r\n")

	h, n, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n == 0 {
		t.Fatalf("expected bytes read > 0")
	}

	host, ok := h.Get("Host")
	if !ok || host != "example.com" {
		t.Fatalf("expected Host header")
	}

	ua, ok := h.Get("User-Agent")
	if !ok || ua != "test" {
		t.Fatalf("expected User-Agent")
	}
}

func TestParseMalformedHeaders(t *testing.T) {
	raw := []byte("BadHeader\r\n\r\n")

	_, _, err := Parse(raw)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMultipleValues(t *testing.T) {
	raw := []byte("Accept: text/html\r\nAccept: text/plain\r\n\r\n")

	h, _, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values := h.Values("Accept")
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
}

func TestCanonicalFormat(t *testing.T) {
	h := New()
	h.Add("content-type", "text/plain")
	h.Add("user-agent", "abc")

	out := string(Format(h))

	if !contains(out, "Content-Type: text/plain\r\n") {
		t.Fatalf("canonicalization failed")
	}
	if !contains(out, "User-Agent: abc\r\n") {
		t.Fatalf("canonicalization failed")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) &&
		(len(s) == len(sub) && s == sub || (len(s) > len(sub) && (s[:len(sub)] == sub || contains(s[1:], sub))))
}
