package cmd

import "testing"

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
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
