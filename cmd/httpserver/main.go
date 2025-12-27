package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pier/internal/request"
	"pier/internal/response"
	"pier/internal/server"
	"strings"
	"syscall"
)

func html(title, heading, msg string) []byte {
	b := make([]byte, 0, 256)
	b = fmt.Appendf(
		b,
		`<html>
<head>
	<title>%s</title>
</head>
<body>
	<h1>%s</h1>
	<p>%s</p>
</body>
</html>`,
		title,
		heading,
		msg,
	)
	return b
}

func logging(next server.Handler) server.Handler {
	return func(w *response.Writer, r *request.Request) {
		log.Printf("%s %s", r.Method, r.Target)
		next(w, r)
	}
}

func handleHttpBin(w *response.Writer, path string) {
	res, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		body := html("500", "Internal Server Error", "Proxy failed")
		headers := response.DefaultHeaders(len(body), "text/html", false)
		w.Write(response.StatusInternalServerError, headers, body)
		return
	}
	defer res.Body.Close()

	var all []byte
	buf := make([]byte, 1024)

	w.WriteRaw([]byte("HTTP/1.1 200 OK\r\n"))
	w.WriteRaw([]byte("Transfer-Encoding: chunked\r\n"))
	w.WriteRaw([]byte("Content-Type: text/plain\r\n"))
	w.WriteRaw([]byte("Trailer: X-Content-SHA256, X-Content-Length\r\n\r\n"))

	for {
		n, err := res.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			all = append(all, chunk...)
			w.WriteChunk(chunk)
		}

		if err != nil {
			break
		}
	}

	w.EndChunked()

	sum := sha256.Sum256(all)

	w.WriteRaw([]byte("X-Content-SHA256: "))
	w.WriteRaw([]byte(hex.EncodeToString(sum[:])))
	w.WriteRaw([]byte("\r\n"))

	buf = make([]byte, 0, 64)
	buf = fmt.Appendf(buf, "X-Content-Length: %d\r\n\r\n", len(all))
	w.WriteRaw(buf)
}

func main() {
	handler := func(w *response.Writer, r *request.Request) {
		switch {
		case r.Target == "/bad":
			body := html("400", "Bad Request", "Invalid request")
			headers := response.DefaultHeaders(len(body), "text/html", false)
			w.Write(response.StatusBadRequest, headers, body)

		case r.Target == "/video":
			data, err := os.ReadFile("assets/video.mp4")
			if err != nil {
				body := html("500", "Internal Server Error", "Unable to load resource")
				headers := response.DefaultHeaders(len(body), "text/html", false)
				w.Write(response.StatusInternalServerError, headers, body)
				return
			}

			headers := response.DefaultHeaders(len(data), "video/mp4", true)
			w.Write(response.StatusOK, headers, data)

		case strings.HasPrefix(r.Target, "/httpbin/"):
			target := strings.TrimPrefix(r.Target, "/httpbin/")
			handleHttpBin(w, target)

		default:
			body := html("200", "OK", "Request successful")
			headers := response.DefaultHeaders(len(body), "text/html", true)
			w.Write(response.StatusOK, headers, body)
		}
	}

	s, err := server.Serve(42069, handler, logging)

	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	log.Println("Server running on port 42069")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	log.Println("Server shutting down")
}
