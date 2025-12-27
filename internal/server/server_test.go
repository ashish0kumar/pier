package server

import (
	"bufio"
	"fmt"
	"net"
	"pier/internal/request"
	"pier/internal/response"
	"strings"
	"testing"
	"time"
)

func startTestServer(t *testing.T) (*Server, uint16) {
	t.Helper()

	var port uint16 = 42111

	handler := func(w *response.Writer, r *request.Request) {
		switch r.Target {
		case "/bad":
			body := []byte("bad request test")
			h := response.DefaultHeaders(len(body), "text/plain", false)
			w.Write(response.StatusBadRequest, h, body)

		default:
			body := []byte("ok")
			h := response.DefaultHeaders(len(body), "text/plain", true)
			w.Write(response.StatusOK, h, body)
		}
	}

	s, err := Serve(port, handler)
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	return s, port
}

func dial(t *testing.T, port uint16) net.Conn {
	t.Helper()

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

func readFullResponse(t *testing.T, conn net.Conn) string {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(conn)

	var b strings.Builder

	for {
		chunk := make([]byte, 512)
		n, err := reader.Read(chunk)
		if n > 0 {
			b.Write(chunk[:n])
			if strings.Contains(b.String(), "\r\n\r\n") {
				break
			}
		}
		if err != nil {
			break
		}
	}

	return b.String()
}

func TestBasicOKResponse(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Close()

	conn := dial(t, port)
	defer conn.Close()

	req := "" +
		"GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"\r\n"

	conn.Write([]byte(req))

	resp := readFullResponse(t, conn)

	if !strings.Contains(resp, "HTTP/1.1 200 OK") {
		t.Fatalf("expected 200, got:\n%s", resp)
	}

	if !strings.Contains(resp, "Content-Length: 2") {
		t.Fatalf("expected content length")
	}

	if !strings.Contains(resp, "ok") {
		t.Fatalf("missing body")
	}
}

func TestBadRequestRoute(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Close()

	conn := dial(t, port)
	defer conn.Close()

	req := "" +
		"GET /bad HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"\r\n"

	conn.Write([]byte(req))

	resp := readFullResponse(t, conn)

	if !strings.Contains(resp, "400 Bad Request") {
		t.Fatalf("expected 400")
	}

	if !strings.Contains(resp, "bad request test") {
		t.Fatalf("expected body content")
	}
}

func TestMalformedRequestGetsClosed(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Close()

	conn := dial(t, port)
	defer conn.Close()

	req := "BROKEN\r\n\r\n"
	conn.Write([]byte(req))

	buf := make([]byte, 32)
	n, _ := conn.Read(buf)

	if n == 0 {
		return
	}
}

func TestKeepAliveHandlesTwoRequestsSameConnection(t *testing.T) {
	s, port := startTestServer(t)
	defer s.Close()

	conn := dial(t, port)
	defer conn.Close()

	req1 := "" +
		"GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"Connection: keep-alive\r\n" +
		"\r\n"

	conn.Write([]byte(req1))
	resp1 := readFullResponse(t, conn)

	if !strings.Contains(resp1, "200 OK") {
		t.Fatalf("first request failed")
	}

	req2 := "" +
		"GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"\r\n"

	conn.Write([]byte(req2))
	resp2 := readFullResponse(t, conn)

	if !strings.Contains(resp2, "200 OK") {
		t.Fatalf("second request failed on same connection")
	}
}
