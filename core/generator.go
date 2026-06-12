package core

import (
	"fmt"
	"strings"
)

func (r *Registry) GenerateCaddyfile() (string, error) {
	rules, err := r.List()
	if err != nil {
		return "", err
	}

	httpPort, httpsPort, err := r.GetPorts()
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	sb.WriteString("{\n")
	sb.WriteString("\tadmin localhost:2019\n")
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

		sb.WriteString(fmt.Sprintf("%s {\n", rule.PublicPattern))
		sb.WriteString(fmt.Sprintf("\treverse_proxy %s {\n", rule.InternalTarget))
		sb.WriteString(fmt.Sprintf("\t\theader_up Host %s\n", hostHeader))
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"http://%s(.*)\" \"https://%s$1\"\n", hostHeader, rule.PublicPattern))
		sb.WriteString(fmt.Sprintf("\t\theader_down Location \"https://%s(.*)\" \"https://%s$1\"\n", hostHeader, rule.PublicPattern))
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")
	}

	return sb.String(), nil
}
