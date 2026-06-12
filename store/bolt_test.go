package store

import (
	"os"
	"testing"

	"github.com/korfra/veyra/model"
)

func TestStore(t *testing.T) {
	tempDbFile := "test_veyra.db"
	defer os.Remove(tempDbFile)

	s, err := New(tempDbFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	rule := model.Rule{
		PublicPattern:  "test.localhost",
		InternalTarget: "localhost:3000",
	}
	err = s.Save(rule)
	if err != nil {
		t.Fatalf("failed to save rule: %v", err)
	}

	got, err := s.Get("test.localhost")
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}
	if got == nil || got.InternalTarget != "localhost:3000" {
		t.Errorf("got rule = %v; want target localhost:3000", got)
	}

	rules, err := s.List()
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
	}
	if len(rules) != 1 || rules[0].PublicPattern != "test.localhost" {
		t.Errorf("listed rules = %v; want 1 rule for test.localhost", rules)
	}

	err = s.SetSetting("http_port", "8080")
	if err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}
	val, err := s.GetSetting("http_port")
	if err != nil {
		t.Fatalf("failed to get setting: %v", err)
	}
	if val != "8080" {
		t.Errorf("got setting = %q; want 8080", val)
	}

	err = s.Delete("test.localhost")
	if err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}
	gotDeleted, err := s.Get("test.localhost")
	if err != nil {
		t.Fatalf("failed to get deleted rule: %v", err)
	}
	if gotDeleted != nil {
		t.Errorf("got deleted rule = %v; want nil", gotDeleted)
	}
}

func TestStoreErrors(t *testing.T) {
	s, err := New("/nonexistent-dir/veyra.db")
	if err == nil {
		if s != nil {
			s.Close()
		}
		t.Error("expected error opening db at nonexistent path, got nil")
	}

	var emptyStore Store
	err = emptyStore.Close()
	if err != nil {
		t.Errorf("expected nil error closing empty store, got: %v", err)
	}
}
