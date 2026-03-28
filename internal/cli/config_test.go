package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	configpkg "github.com/aimony/mihosh/internal/infrastructure/config"
)

func TestRenderConfigShow(t *testing.T) {
	cfg := &configpkg.Config{
		APIAddress:   "http://127.0.0.1:9090",
		Secret:       "secret-token",
		TestURL:      "http://www.gstatic.com/generate_204",
		Timeout:      5000,
		ProxyAddress: "http://127.0.0.1:7890",
	}

	tests := []struct {
		name     string
		format   outputFormat
		contains []string
	}{
		{
			name:   "JSON format",
			format: outputFormatJSON,
			contains: []string{
				`"api_address": "http://127.0.0.1:9090"`,
				`"secret": "sec****ken"`,
				`"config_file":`,
				`config.yaml`,
			},
		},
		{
			name:   "Table format",
			format: outputFormatTable,
			contains: []string{
				"KEY",
				"VALUE",
				"API_ADDRESS",
				"http://127.0.0.1:9090",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := renderConfigShow(&out, cfg, "C:\\mihosh\\config.yaml", tt.format)
			assert.NoError(t, err)

			output := out.String()
			for _, c := range tt.contains {
				assert.Contains(t, output, c)
			}
		})
	}
}
