# pier

> **HTTP/1.1 server** written from scratch in Go

- request parsing
- header parsing + canonicalization
- Content-Length handling
- Transfer-Encoding: chunked (request + response)
- persistent connections (keep-alive)
- timeouts
- middleware pipeline
- streaming proxy to httpbin
- graceful error handling

---

## Features

### Core HTTP
- HTTP/1.1 request line parsing
- header parsing (case-insensitive, multi-value)
- body parsing with:
  - `Content-Length`
  - `Transfer-Encoding: chunked`
- timeout protection

### Response Engine
- manual status line writing
- canonical header writing
- Content-Length responses
- chunked responses (`WriteChunk` + trailers)
- binary-safe body streaming

### Connection Handling
- persistent connections
- `Connection: close` respected
- read deadlines
- graceful close

### Middleware
- chainable middleware system
- built-in logging example

### Extras
- `/bad` returns 400 example
- `/video` streams a binary file
- `/httpbin/...` proxy endpoint
  - streamed chunked response
  - SHA256 + length trailers
- unit + integration tests

---

## Project Structure

```
cmd/
├── httpserver -> main HTTP server
└── tcplistener -> debug raw TCP inspector

internal/
├── headers -> header parsing + formatting
├── request -> HTTP request parsing
├── response -> response writer + chunking
└── server -> TCP server + middleware
```

---

## Running

```bash
# Start the server (port 42069)
go run ./cmd/httpserver
```

### Try it

```bash
# Basic request
curl -v http://localhost:42069/

# 400 Bad Request example
curl -v http://localhost:42069/bad

# Video streaming (requires assets/video.mp4)
curl -v http://localhost:42069/video --output out.mp4

# Chunked proxy (with trailers)
curl -v http://localhost:42069/httpbin/get
```

---

## Testing

```bash
# Unit tests
go test ./internal/headers
go test ./internal/request
go test ./internal/response

# Integration tests
go test ./internal/server

# Full sweep
go test ./...
```

---

## Endpoints

| Path | Description |
|------|------------|
| `/` | basic 200 HTML |
| `/bad` | returns 400 response |
| `/video` | streams `assets/video.mp4` |
| `/httpbin/...` | proxies to httpbin with chunked streaming + trailers |

<br><br>

<p align="center">
	<img src="https://raw.githubusercontent.com/catppuccin/catppuccin/main/assets/footers/gray0_ctp_on_line.svg?sanitize=true" />
</p>

<p align="center">
    <i><code>&copy 2025-present <a href="https://github.com/ashish0kumar">Ashish Kumar</a></code></i>
</p>

<div align="center">
<a href="https://github.com/ashish0kumar/pier/blob/main/LICENSE"><img src="https://img.shields.io/github/license/ashish0kumar/pier?style=for-the-badge&color=CBA6F7&logoColor=cdd6f4&labelColor=302D41" alt="LICENSE"></a>&nbsp;&nbsp;
</div>
