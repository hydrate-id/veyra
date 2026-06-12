package caddy

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(BodyRewriter{})
	httpcaddyfile.RegisterHandlerDirective("body_rewriter", parseCaddyfile)
}

type BodyRewriter struct {
	Search  string `json:"search,omitempty"`
	Replace string `json:"replace,omitempty"`
}

func (BodyRewriter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.body_rewriter",
		New: func() caddy.Module { return new(BodyRewriter) },
	}
}

func (m *BodyRewriter) Provision(ctx caddy.Context) error {
	return nil
}

func (m *BodyRewriter) Validate() error {
	return nil
}

func (m *BodyRewriter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&m.Search, &m.Replace) {
			return d.ArgErr()
		}
	}
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m BodyRewriter
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return &m, err
}

type responseBuffer struct {
	http.ResponseWriter
	body        *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func (rb *responseBuffer) WriteHeader(statusCode int) {
	if !rb.wroteHeader {
		rb.statusCode = statusCode
		rb.wroteHeader = true
	}
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	return rb.body.Write(b)
}

func (rb *responseBuffer) Header() http.Header {
	return rb.ResponseWriter.Header()
}

func (m *BodyRewriter) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	resolvedReplace := repl.ReplaceKnown(m.Replace, "")

	buf := &bytes.Buffer{}
	rb := &responseBuffer{
		ResponseWriter: w,
		body:           buf,
	}

	err := next.ServeHTTP(rb, r)
	if err != nil {
		return err
	}

	bodyBytes := buf.Bytes()
	contentType := w.Header().Get("Content-Type")

	isText := false
	if contentType == "" {
		isText = true
	} else {
		ct := strings.ToLower(contentType)
		if strings.Contains(ct, "text/") ||
			strings.Contains(ct, "json") ||
			strings.Contains(ct, "javascript") ||
			strings.Contains(ct, "xml") {
			isText = true
		}
	}

	if isText && len(bodyBytes) > 0 {
		bodyStr := string(bodyBytes)
		bodyStr = strings.ReplaceAll(bodyStr, m.Search, resolvedReplace)
		bodyBytes = []byte(bodyStr)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))

	if rb.wroteHeader {
		w.WriteHeader(rb.statusCode)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_, err = w.Write(bodyBytes)
	return err
}

var (
	_ caddy.Provisioner           = (*BodyRewriter)(nil)
	_ caddy.Validator             = (*BodyRewriter)(nil)
	_ caddyhttp.MiddlewareHandler = (*BodyRewriter)(nil)
	_ caddyfile.Unmarshaler       = (*BodyRewriter)(nil)
)
