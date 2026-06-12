package caddy

import (
	"net/http"
	"testing"
	"time"
)

func TestAdaptCaddyfile(t *testing.T) {
	caddyfile := `
{
	admin localhost:2020
}
localhost:8081 {
	respond "OK"
}
`
	jsonBytes, err := AdaptCaddyfile(caddyfile)
	if err != nil {
		t.Fatalf("failed to adapt Caddyfile: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON bytes")
	}
}

func TestStartStopReload(t *testing.T) {
	caddyfileDefaultAdmin := `
{
	admin localhost:2019
	http_port 8082
	https_port 8445
}
localhost:8082 {
	respond "OK"
}
`
	err := Start(caddyfileDefaultAdmin)
	if err != nil {
		t.Fatalf("failed to start Caddy: %v", err)
	}
	defer func() {
		_ = Stop()
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8082")
	if err != nil {
		t.Fatalf("failed to query started Caddy: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status code %d; want 200", resp.StatusCode)
	}

	err = Reload(caddyfileDefaultAdmin, 2019)
	if err != nil {
		t.Errorf("failed to reload Caddy: %v", err)
	}
}

func TestCaddyErrors(t *testing.T) {
	malformed := `
{
	invalid_directive
}
`
	_, err := AdaptCaddyfile(malformed)
	if err == nil {
		t.Error("expected error adapting malformed Caddyfile, got nil")
	}

	err = Start(malformed)
	if err == nil {
		t.Error("expected error starting Caddy with malformed Caddyfile, got nil")
	}

	_ = Stop()
	time.Sleep(1 * time.Second)

	// Since Caddy is stopped, Reloading on port 2021 (which was never started) must fail!
	err = Reload(`
{
	admin localhost:2021
}
`, 2021)
	if err == nil {
		t.Error("expected error reloading Caddy when stopped, got nil")
	}
}

func TestReloadErrorResponse(t *testing.T) {
	caddyfileValid := `
{
	admin localhost:2023
	http_port 8085
}
localhost:8085 {
	respond "OK"
}
`
	err := Start(caddyfileValid)
	if err != nil {
		t.Fatalf("failed to start Caddy: %v", err)
	}
	defer func() {
		_ = Stop()
	}()

	caddyfileInvalidLoad := `
{
	admin localhost:2023
	http_port 80
}
localhost:80 {
	respond "OK"
}
`
	err = Reload(caddyfileInvalidLoad, 2023)
	if err == nil {
		t.Error("expected error reloading Caddy with port 80 binding, got nil")
	}
}
