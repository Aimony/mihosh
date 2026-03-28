package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    outputFormat
		wantErr bool
	}{
		{name: "json", input: "json", want: outputFormatJSON},
		{name: "table", input: "table", want: outputFormatTable},
		{name: "plain", input: "plain", want: outputFormatPlain},
		{name: "invalid", input: "yaml", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOutputFormat(tc.input)
			if tc.wantErr {
				assert.Error(t, err, "expected error for input %q", tc.input)
				return
			}
			assert.NoError(t, err, "unexpected error for input %q", tc.input)
			assert.Equal(t, tc.want, got, "expected %q, got %q", tc.want, got)
		})
	}
}
