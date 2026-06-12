package core

import (
	"fmt"
	"strings"
)

func getWildcardLabelIndex(pattern string) int {
	parts := strings.Split(pattern, ".")
	for i, part := range parts {
		if part == "*" {
			return len(parts) - 1 - i
		}
	}
	return -1
}

// GenerateCaddyfile generates a Caddyfile from the registered rules.
func (r *Registry) GenerateCaddyfile() (string, error) {
	rules, err := r.List()
	if err != nil {
		return "", err
	}

	httpPort, httpsPort, err := r.GetPorts()
	if err != nil {
		return "", err
	}

	adminPort, err := r.GetAdminPort()
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Add global block to explicitly configure the admin API on localhost and custom ports
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("\tadmin localhost:%d\n", adminPort))
	if httpPort != 80 {
		sb.WriteString(fmt.Sprintf("\thttp_port %d\n", httpPort))
	}
	if httpsPort != 443 {
		sb.WriteString(fmt.Sprintf("\thttps_port %d\n", httpsPort))
	}
	sb.WriteString("}\n\n")

	for _, rule := range rules {
		internalHost := rule.InternalTarget
		hostHeader := internalHost
		if strings.HasPrefix(hostHeader, "http://") {
			hostHeader = strings.TrimPrefix(hostHeader, "http://")
		} else if strings.HasPrefix(hostHeader, "https://") {
			hostHeader = strings.TrimPrefix(hostHeader, "https://")
		}

		targetForCaddy := rule.InternalTarget
		hostHeaderForCaddy := hostHeader

		// Handle wildcards (capture and flatten)
		if wildcardIdx := getWildcardLabelIndex(rule.PublicPattern); wildcardIdx != -1 {
			if strings.Contains(rule.InternalTarget, "*") {
				placeholder := fmt.Sprintf("{labels.%d}", wildcardIdx)
				targetForCaddy = strings.ReplaceAll(rule.InternalTarget, "*", placeholder)
				hostHeaderForCaddy = strings.ReplaceAll(hostHeader, "*", placeholder)
			}
		}

		sb.WriteString(fmt.Sprintf("%s {\n", rule.PublicPattern))
		sb.WriteString(fmt.Sprintf("\treverse_proxy %s {\n", targetForCaddy))
		sb.WriteString(fmt.Sprintf("\t\theader_up Host %s\n", hostHeaderForCaddy))
		// Rewrite redirect location response headers dynamically using {host}
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"http://%s(.*)\" \"https://{host}$1\"\n", hostHeaderForCaddy))
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"https://%s(.*)\" \"https://{host}$1\"\n", hostHeaderForCaddy))
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}
