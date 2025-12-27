headers
package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers struct {
	m map[string][]string
}

var (
	crlf         = []byte("\r\n")
	ErrMalformed = errors.New("malformed headers")
)

func New() *Headers {
	return &Headers{m: map[string][]string{}}
}

func (h *Headers) Get(name string) (string, bool) {
	v := h.m[strings.ToLower(name)]
	if len(v) == 0 {
		return "", false
	}
	return v[0], true
}

func (h *Headers) Values(name string) []string {
	return h.m[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	h.m[strings.ToLower(name)] = []string{value}
}

func (h *Headers) Add(name, value string) {
	k := strings.ToLower(name)
	h.m[k] = append(h.m[k], value)
}

func (h *Headers) Delete(name string) {
	delete(h.m, strings.ToLower(name))
}

func Parse(b []byte) (*Headers, int, error) {
	h := New()
	read := 0

	for {
		idx := bytes.Index(b[read:], crlf)
		if idx < 0 {
			return nil, 0, nil
		}

		if idx == 0 {
			read += 2
			break
		}

		line := b[read : read+idx]
		read += idx + 2

		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return nil, 0, ErrMalformed
		}

		name := strings.TrimSpace(string(parts[0]))
		value := strings.TrimSpace(string(parts[1]))

		if name == "" {
			return nil, 0, ErrMalformed
		}

		h.Add(name, value)
	}

	return h, read, nil
}

func Format(h *Headers) []byte {
	var buf bytes.Buffer
	for k, vals := range h.m {
		name := canonical(k)
		for _, v := range vals {
			buf.WriteString(name)
			buf.WriteString(": ")
			buf.WriteString(v)
			buf.Write(crlf)
		}
	}
	buf.Write(crlf)
	return buf.Bytes()
}

func canonical(k string) string {
	parts := strings.Split(k, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "-")
}

func (h *Headers) Canonical() map[string][]string {
	out := map[string][]string{}
	for k, v := range h.m {
		out[canonical(k)] = v
	}
	return out
}
