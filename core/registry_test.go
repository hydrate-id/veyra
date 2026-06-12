package core

import (
	"os"
	"strings"
	"testing"

	"github.com/korfra/veyra/store"
)

func TestRegistryAndGenerator(t *testing.T) {
	tempDbFile := "test_registry.db"
	defer os.Remove(tempDbFile)

	s, err := store.New(tempDbFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	r := New(s)

	hp, sp, err := r.GetPorts()
	if err != nil {
		t.Errorf("failed to get default ports: %v", err)
	}
	if hp != 8080 || sp != 8443 {
		t.Errorf("got default ports %d, %d; want 8080, 8443", hp, sp)
	}

	err = r.SetPorts(8080, 8443)
	if err != nil {
		t.Fatalf("failed to set ports: %v", err)
	}
	hp, sp, err = r.GetPorts()
	if err != nil {
		t.Errorf("failed to get ports: %v", err)
	}
	if hp != 8080 || sp != 8443 {
		t.Errorf("got ports %d, %d; want 8080, 8443", hp, sp)
	}

	err = r.Add("example.com", "localhost:8000")
	if err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	err = r.Add("secure.example.com", "https://localhost:8001")
	if err != nil {
		t.Fatalf("failed to add secure rule: %v", err)
	}

	rule, err := r.Get("example.com")
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}
	if rule == nil || rule.InternalTarget != "localhost:8000" {
		t.Errorf("got rule %v; want localhost:8000", rule)
	}

	list, err := r.List()
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("got list length %d; want 2", len(list))
	}

	caddyfile, err := r.GenerateCaddyfile()
	if err != nil {
		t.Fatalf("failed to generate Caddyfile: %v", err)
	}
	if !strings.Contains(caddyfile, "example.com") {
		t.Errorf("expected Caddyfile to contain example.com: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "http_port 8080") {
		t.Errorf("expected Caddyfile to contain http_port 8080: %s", caddyfile)
	}
	if strings.Contains(caddyfile, "https_port 8443") {
		t.Errorf("expected Caddyfile to NOT contain https_port 8443 when disabled: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "header_up X-Forwarded-Host localhost:8000") {
		t.Errorf("expected Caddyfile to contain header_up X-Forwarded-Host localhost:8000, got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "header_up -Accept-Encoding") {
		t.Errorf("expected Caddyfile to contain header_up -Accept-Encoding, got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "header_down Set-Cookie \"localhost:8000\" \"{host}\"") {
		t.Errorf("expected Caddyfile to contain header_down Set-Cookie \"localhost:8000\" \"{host}\", got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "order body_rewriter after encode") {
		t.Errorf("expected Caddyfile to contain order body_rewriter after encode, got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "body_rewriter localhost:8000 {host}") {
		t.Errorf("expected Caddyfile to contain body_rewriter localhost:8000 {host}, got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "header_up X-Forwarded-Port 80") {
		t.Errorf("expected Caddyfile to contain header_up X-Forwarded-Port 80, got: %s", caddyfile)
	}
	if strings.Contains(caddyfile, "https://example.com") {
		t.Errorf("expected Caddyfile to NOT contain https://example.com when disabled, got: %s", caddyfile)
	}

	err = r.SetHTTPSEnabled(true)
	if err != nil {
		t.Fatalf("failed to enable HTTPS: %v", err)
	}

	caddyfileHTTPS, err := r.GenerateCaddyfile()
	if err != nil {
		t.Fatalf("failed to generate Caddyfile with HTTPS: %v", err)
	}
	if !strings.Contains(caddyfileHTTPS, "https_port 8443") {
		t.Errorf("expected Caddyfile with HTTPS to contain https_port 8443: %s", caddyfileHTTPS)
	}
	if !strings.Contains(caddyfileHTTPS, "https://example.com") {
		t.Errorf("expected Caddyfile with HTTPS to contain https://example.com: %s", caddyfileHTTPS)
	}
	if !strings.Contains(caddyfileHTTPS, "header_up X-Forwarded-Proto https") {
		t.Errorf("expected Caddyfile with HTTPS to contain header_up X-Forwarded-Proto https, got: %s", caddyfileHTTPS)
	}
	if !strings.Contains(caddyfileHTTPS, "header_up X-Forwarded-Port 443") {
		t.Errorf("expected Caddyfile with HTTPS to contain header_up X-Forwarded-Port 443, got: %s", caddyfileHTTPS)
	}

	err = r.Remove("example.com")
	if err != nil {
		t.Fatalf("failed to remove rule: %v", err)
	}
	err = r.Remove("secure.example.com")
	if err != nil {
		t.Fatalf("failed to remove secure rule: %v", err)
	}
	rule, err = r.Get("example.com")
	if err != nil {
		t.Fatalf("failed to check removed rule: %v", err)
	}
	if rule != nil {
		t.Errorf("expected rule to be nil, got %v", rule)
	}
	err = r.SetAdminPort(2022)
	if err != nil {
		t.Fatalf("failed to set admin port: %v", err)
	}
	port, err := r.GetAdminPort()
	if err != nil {
		t.Fatalf("failed to get admin port: %v", err)
	}
	if port != 2022 {
		t.Errorf("got admin port %d; want 2022", port)
	}

	err = r.Add("backup.localhost", "localhost:5000")
	if err != nil {
		t.Fatalf("failed to add rule for backup: %v", err)
	}

	backup, err := r.ExportBackup()
	if err != nil {
		t.Fatalf("failed to export backup: %v", err)
	}
	if len(backup.Rules) != 1 || backup.Rules[0].PublicPattern != "backup.localhost" {
		t.Errorf("unexpected backup rules: %v", backup.Rules)
	}
	if backup.Settings["admin_port"] != "2022" {
		t.Errorf("expected admin_port to be 2022 in backup settings, got: %s", backup.Settings["admin_port"])
	}

	err = r.Remove("backup.localhost")
	if err != nil {
		t.Fatalf("failed to remove rule: %v", err)
	}

	err = r.ImportBackup(backup)
	if err != nil {
		t.Fatalf("failed to import backup: %v", err)
	}

	rule, err = r.Get("backup.localhost")
	if err != nil {
		t.Fatalf("failed to get imported rule: %v", err)
	}
	if rule == nil || rule.InternalTarget != "localhost:5000" {
		t.Errorf("imported rule mismatch: %v", rule)
	}

	err = r.Remove("backup.localhost")
	if err != nil {
		t.Fatalf("failed to clean up imported rule: %v", err)
	}

	err = s.SetSetting("admin_port", "not-a-number")
	if err != nil {
		t.Fatalf("failed to set invalid admin port setting: %v", err)
	}
	port, err = r.GetAdminPort()
	if err != nil {
		t.Errorf("expected no error when admin port is invalid, got: %v", err)
	}
	if port != 2019 {
		t.Errorf("got admin port %d; want default 2019 for invalid setting", port)
	}
}

func TestRegistryErrors(t *testing.T) {
	tempDbFile := "test_registry_errors.db"
	defer os.Remove(tempDbFile)

	s, err := store.New(tempDbFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	r := New(s)

	_ = s.Close()

	_, _, err = r.GetPorts()
	if err == nil {
		t.Error("expected error getting ports from closed store, got nil")
	}

	_, err = r.GetAdminPort()
	if err == nil {
		t.Error("expected error getting admin port from closed store, got nil")
	}

	_, err = r.GenerateCaddyfile()
	if err == nil {
		t.Error("expected error generating Caddyfile from closed store, got nil")
	}
}

func TestValidationAndWildcards(t *testing.T) {
	tempDbFile := "test_val_wildcard.db"
	defer os.Remove(tempDbFile)

	s, err := store.New(tempDbFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	r := New(s)

	tests := []struct {
		public   string
		internal string
		wantErr  bool
	}{
		{"", "localhost:3000", true},
		{"example.com", "", true},
		{"http://example.com", "localhost", true},
		{"example.com:80", "localhost", true},
		{"invalid_host!!", "localhost", true},
		{"example.com", "localhost/path", true},
		{"example.com", "http://localhost", false},
		{"example2.com", "localhost:3000", false},
	}

	for _, tt := range tests {
		err := r.Add(tt.public, tt.internal)
		if (err != nil) != tt.wantErr {
			t.Errorf("Add(%q, %q) error = %v; wantErr %v", tt.public, tt.internal, err, tt.wantErr)
		}
	}

	err = r.Add("example.com", "localhost:4000")
	if err == nil {
		t.Error("expected conflict error adding duplicate public pattern, got nil")
	}

	err = r.Add("*.zahra.syntopia.dev", "*.zahra.test")
	if err != nil {
		t.Fatalf("failed to add wildcard capture: %v", err)
	}

	err = r.Add("*.demo.syntopia.dev", "plane.zahra.test")
	if err != nil {
		t.Fatalf("failed to add wildcard flatten: %v", err)
	}

	caddyfile, err := r.GenerateCaddyfile()
	if err != nil {
		t.Fatalf("failed to generate Caddyfile: %v", err)
	}

	if !strings.Contains(caddyfile, "{labels.3}.zahra.test") {
		t.Errorf("expected Caddyfile to contain {labels.3}.zahra.test, got: %s", caddyfile)
	}
	if !strings.Contains(caddyfile, "plane.zahra.test") {
		t.Errorf("expected Caddyfile to contain plane.zahra.test, got: %s", caddyfile)
	}
}
