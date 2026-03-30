package components

import (
	"fmt"
	"strings"
)

func renderKVLine(key, value string, s detailStyles) string {
	return fmt.Sprintf("%s %s", s.Label.Render(key+"："), s.Value.Render(value))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatEndpoint(ip, port string) string {
	if ip == "" {
		return "-"
	}
	if port == "" {
		return ip
	}
	return ip + ":" + port
}
