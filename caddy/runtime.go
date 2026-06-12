package caddy

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	_ "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func AdaptCaddyfile(caddyfileContent string) ([]byte, error) {
	adapter := caddyconfig.GetAdapter("caddyfile")
	if adapter == nil {
		return nil, fmt.Errorf("caddyfile adapter not found")
	}

	jsonBytes, warnings, err := adapter.Adapt([]byte(caddyfileContent), map[string]any{
		"filename": "Caddyfile",
	})
	if err != nil {
		return nil, err
	}

	_ = warnings

	return jsonBytes, nil
}

func Start(caddyfileContent string) error {
	jsonBytes, err := AdaptCaddyfile(caddyfileContent)
	if err != nil {
		return fmt.Errorf("failed to adapt Caddyfile: %w", err)
	}

	err = caddy.Load(jsonBytes, true)
	if err != nil {
		return fmt.Errorf("failed to load caddy config: %w", err)
	}

	return nil
}

func Stop() error {
	return caddy.Stop()
}

func Reload(caddyfileContent string, adminPort int) error {
	jsonBytes, err := AdaptCaddyfile(caddyfileContent)
	if err != nil {
		return fmt.Errorf("failed to adapt Caddyfile: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/load", adminPort), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to send reload request to Caddy admin API: %w (is 'veyra start' running?)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("caddy admin API returned status %s", resp.Status)
	}

	return nil
}
