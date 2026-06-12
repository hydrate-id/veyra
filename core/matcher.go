package core

import (
	"regexp"
	"strings"

	"github.com/korfra/veyra/model"
)

func Match(host string, rules []model.Rule) (string, bool) {

	for _, r := range rules {

		if !isWildcard(r.PublicPattern) {
			if host == r.PublicPattern {
				return r.InternalTarget, true
			}
			continue
		}

		regex := patternToRegex(r.PublicPattern)
		re := regexp.MustCompile(regex)

		if re.MatchString(host) {

			if strings.Contains(r.InternalTarget, "*") {
				matches := re.FindStringSubmatch(host)
				return strings.Replace(r.InternalTarget, "*", matches[1], 1), true
			}

			return r.InternalTarget, true
		}
	}

	return "", false
}
