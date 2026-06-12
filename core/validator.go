package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/korfra/veyra/model"
)

var hostRegex = regexp.MustCompile(`^(\*|[a-zA-Z0-9][-a-zA-Z0-9]*[a-zA-Z0-9]?)+(\.[a-zA-Z0-9][-a-zA-Z0-9]*[a-zA-Z0-9]?)*$`)

func ValidateRule(rule model.Rule, existingRules []model.Rule) error {
	public := strings.TrimSpace(rule.PublicPattern)
	internal := strings.TrimSpace(rule.InternalTarget)

	if public == "" {
		return errors.New("public pattern cannot be empty")
	}
	if internal == "" {
		return errors.New("internal target cannot be empty")
	}

	if strings.Contains(public, "://") {
		return errors.New("public pattern must not contain protocol scheme (e.g., http://)")
	}

	if strings.Contains(public, ":") {
		return errors.New("public pattern must not contain a port number")
	}

	if !hostRegex.MatchString(public) {
		return fmt.Errorf("invalid hostname format for public pattern: %s", public)
	}

	for _, existing := range existingRules {
		if existing.PublicPattern == public {
			return fmt.Errorf("conflict: rule for public pattern '%s' already exists", public)
		}
	}

	targetHost := internal
	if strings.Contains(targetHost, "://") {
		parts := strings.SplitN(targetHost, "://", 2)
		targetHost = parts[1]
	}
	if strings.Contains(targetHost, "/") {
		return errors.New("internal target must not contain a path (e.g., /path)")
	}

	return nil
}
