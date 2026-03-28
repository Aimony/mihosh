package cmd

import (
	"bytes"
	"strings"
	"testing"

	configpkg "github.com/aimony/mihosh/internal/infrastructure/config"
)

func TestRenderConfigShowJSON(t *testing.T) {
	cfg := &configpkg.Config{
		APIAddress:   "http://127.0.0.1:9090",
		Secret:       "secret-token",
		TestURL:      "http://www.gstatic.com/generate_204",
		Timeout:      5000,
		ProxyAddress: "http://127.0.0.1:7890",
	}

	var out bytes.Buffer
	if err := renderConfigShow(&out, cfg, "C:\\mihosh\\config.yaml", outputFormatJSON); err != nil {
		t.Fatalf("renderConfigShow returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"api_address": "http://127.0.0.1:9090"`) {
		t.Fatalf("expected api_address in json output, got:\n%s", output)
	}
	if !strings.Contains(output, `"secret": "sec****ken"`) {
		t.Fatalf("expected masked secret in json output, got:\n%s", output)
	}
	if !strings.Contains(output, `"config_file":`) || !strings.Contains(output, `config.yaml`) {
		t.Fatalf("expected config_file in json output, got:\n%s", output)
	}
}

func TestRenderConfigShowTable(t *testing.T) {
	cfg := &configpkg.Config{
		APIAddress:   "http://127.0.0.1:9090",
		Secret:       "secret-token",
		TestURL:      "http://www.gstatic.com/generate_204",
		Timeout:      5000,
		ProxyAddress: "http://127.0.0.1:7890",
	}

	var out bytes.Buffer
	if err := renderConfigShow(&out, cfg, "C:\\mihosh\\config.yaml", outputFormatTable); err != nil {
		t.Fatalf("renderConfigShow returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "KEY") || !strings.Contains(output, "VALUE") {
		t.Fatalf("expected table header in output, got:\n%s", output)
	}
	if !strings.Contains(output, "API_ADDRESS") || !strings.Contains(output, "http://127.0.0.1:9090") {
		t.Fatalf("expected table row in output, got:\n%s", output)
	}
}
