package tools

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	bodySizeLimit = 1000
)

type HTTPMockHandler struct {
	server   *http.ServeMux
	listener net.Listener
}

type responseLogger struct {
	w       http.ResponseWriter
	out     io.StringWriter
	rspBody bytes.Buffer
}

func newResponseLogger(w http.ResponseWriter, out io.StringWriter) *responseLogger {
	return &responseLogger{
		w:   w,
		out: out,
	}
}

func (r *responseLogger) Header() http.Header {
	return r.w.Header()
}

func (r *responseLogger) Write(b []byte) (int, error) {
	r.rspBody.Write(b)
	return r.w.Write(b)
}

func (r *responseLogger) WriteHeader(statusCode int) {
	_, _ = r.out.WriteString(fmt.Sprintf("status: %d\n", statusCode)) //nolint:errcheck

	for k, v := range r.w.Header() {
		_, _ = r.out.WriteString(fmt.Sprintf("	%s: %s\n", k, v)) //nolint:errcheck 
	}
	r.w.WriteHeader(statusCode)
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		buf.WriteString("-----------------\n")
		buf.WriteString(fmt.Sprintf("request: %s %s\n", r.Method, r.RequestURI))

		for k, v := range r.Header {
			buf.WriteString(fmt.Sprintf("	%s: %s\n", k, v))
		}
		data, _ := io.ReadAll(r.Body) //nolint:errcheck 
		_ = r.Body.Close()
		if len(data) > 0 {
			buf.WriteString("req body: ")
			buf.Write(data)
			buf.WriteString("\n")
		}
		r.Body = io.NopCloser(bytes.NewReader(data))

		lw := newResponseLogger(w, &buf)
		buf.WriteString("response:\n")
		next.ServeHTTP(lw, r)

		if lw.rspBody.Len() > 0 {
			buf.WriteString("rsp body: ")
			if lw.rspBody.Len() > bodySizeLimit {
				buf.Write(lw.rspBody.Bytes()[:bodySizeLimit])
				buf.WriteString("...")
			} else {
				buf.Write(lw.rspBody.Bytes())
			}

			buf.WriteString("\n")
		}
		fmt.Println(buf.String())
	})
}

func NewHTTPMockHandler(schema, address string, cb func(server *http.ServeMux)) (*HTTPMockHandler, error) {
	h := &HTTPMockHandler{
		server: http.NewServeMux(),
	}
	cb(h.server)
	l, err := net.Listen(schema, address)
	if err != nil {
		return nil, err
	}
	h.listener = l
	return h, nil
}

func (h *HTTPMockHandler) Start() error {
	var handler http.Handler
	if strings.ToLower(os.Getenv("GOAT_HTTP_DEBUG")) == "true" {
		handler = loggerMiddleware(h.server)
	} else {
		handler = h.server
	}

	return http.Serve(h.listener, handler) //nolint:gosec 
}

func (h *HTTPMockHandler) Stop() error {
	err := h.listener.Close()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed start http server: %w", err)
	}
	return nil
}
