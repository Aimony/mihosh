package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
)

func TestRenderConnectionsJSON(t *testing.T) {
	resp := model.ConnectionsResponse{
		UploadTotal:   1024,
		DownloadTotal: 2048,
		Connections: []model.Connection{
			{
				Metadata: model.Metadata{
					SourceIP:        "10.0.0.1",
					SourcePort:      "1234",
					DestinationIP:   "1.1.1.1",
					DestinationPort: "443",
				},
				Chains: []string{"Proxy", "HK"},
			},
		},
	}

	var out bytes.Buffer
	if err := renderConnections(&out, resp, outputFormatJSON); err != nil {
		t.Fatalf("renderConnections returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"active_connections": 1`) {
		t.Fatalf("expected active_connections in json output, got:\n%s", output)
	}
	if !strings.Contains(output, `"source_ip": "10.0.0.1"`) {
		t.Fatalf("expected source_ip in json output, got:\n%s", output)
	}
}

func TestRenderConnectionsTable(t *testing.T) {
	resp := model.ConnectionsResponse{
		UploadTotal:   1024,
		DownloadTotal: 2048,
		Connections: []model.Connection{
			{
				Metadata: model.Metadata{
					SourceIP:        "10.0.0.1",
					SourcePort:      "1234",
					DestinationIP:   "1.1.1.1",
					DestinationPort: "443",
				},
				Chains: []string{"HK"},
			},
		},
	}

	var out bytes.Buffer
	if err := renderConnections(&out, resp, outputFormatTable); err != nil {
		t.Fatalf("renderConnections returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "SOURCE") || !strings.Contains(output, "DESTINATION") {
		t.Fatalf("expected table header in output, got:\n%s", output)
	}
	if !strings.Contains(output, "10.0.0.1:1234") || !strings.Contains(output, "1.1.1.1:443") {
		t.Fatalf("expected connection row in output, got:\n%s", output)
	}
}
