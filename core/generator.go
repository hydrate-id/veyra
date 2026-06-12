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

	// Add global block to explicitly configure the admin API on localhost, custom ports, and trusted proxies
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("\tadmin localhost:%d\n", adminPort))
	if httpPort != 80 {
		sb.WriteString(fmt.Sprintf("\thttp_port %d\n", httpPort))
	}
	if httpsPort != 443 {
		sb.WriteString(fmt.Sprintf("\thttps_port %d\n", httpsPort))
	}
	sb.WriteString("\tservers {\n")
	sb.WriteString("\t\ttrusted_proxies static private_ranges\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	for _, rule := range rules {
		internalHost := rule.InternalTarget
		hostHeader := internalHost
		isHTTPS := strings.HasPrefix(internalHost, "https://")

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

		// Explicitly define both HTTP and HTTPS schemes to disable Caddy's automatic HTTP->HTTPS
		// redirects which default to port 443 and bypass Veyra's custom ports.
		sb.WriteString(fmt.Sprintf("http://%s, https://%s {\n", rule.PublicPattern, rule.PublicPattern))
		sb.WriteString(fmt.Sprintf("\treverse_proxy %s {\n", targetForCaddy))
		sb.WriteString(fmt.Sprintf("\t\theader_up Host %s\n", hostHeaderForCaddy))
		// Rewrite redirect location response headers dynamically using {host}
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"http://%s(.*)\" \"https://{host}$1\"\n", hostHeaderForCaddy))
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"https://%s(.*)\" \"https://{host}$1\"\n", hostHeaderForCaddy))

		// If proxying to HTTPS backend, skip TLS verification (common for local Valet CA/Self-signed certs)
		if isHTTPS {
			sb.WriteString("\t\ttransport http {\n")
			sb.WriteString("\t\t\ttls_insecure_skip_verify\n")
			sb.WriteString("\t\t}\n")
		}

		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}
