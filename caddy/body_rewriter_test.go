package caddy

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBodyRewriter(t *testing.T) {
	caddyfile := `
{
	admin localhost:2025
	http_port 8083
	order body_rewriter after encode
}
localhost:8083 {
	body_rewriter "hello" "world"
	respond "hello universe"
}
`
	err := Start(caddyfile)
	if err != nil {
		t.Fatalf("failed to start Caddy: %v", err)
	}
	defer func() {
		_ = Stop()
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8083")
	if err != nil {
		t.Fatalf("failed to query Caddy: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(bodyBytes)
	if !strings.Contains(bodyStr, "world universe") {
		t.Errorf("got response body %q; expected it to contain %q", bodyStr, "world universe")
	}
}
