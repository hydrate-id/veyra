package core

import (
	"testing"

	"github.com/korfra/veyra/model"
)

func TestMatch(t *testing.T) {
	rules := []model.Rule{
		{
			PublicPattern:  "zahra.syntopia.dev",
			InternalTarget: "studio.makeupartis.test",
		},
		{
			PublicPattern:  "*.zahra.syntopia.dev",
			InternalTarget: "*.zahra.test",
		},
		{
			PublicPattern:  "*.demo.syntopia.dev",
			InternalTarget: "plane.zahra.test",
		},
	}

	tests := []struct {
		host           string
		expectedTarget string
		expectedMatch  bool
	}{
		{"zahra.syntopia.dev", "studio.makeupartis.test", true},
		{"api.zahra.syntopia.dev", "api.zahra.test", true},
		{"anything.demo.syntopia.dev", "plane.zahra.test", true},
		{"notmatched.dev", "", false},
	}

	for _, tt := range tests {
		target, matched := Match(tt.host, rules)
		if matched != tt.expectedMatch {
			t.Errorf("Match(%q) matched = %v; want %v", tt.host, matched, tt.expectedMatch)
		}
		if target != tt.expectedTarget {
			t.Errorf("Match(%q) target = %q; want %q", tt.host, target, tt.expectedTarget)
		}
	}
}
