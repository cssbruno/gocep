package util

import "testing"

func TestNormalizeCEP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "already_normalized", input: "01001000", want: "01001000"},
		{name: "hyphen", input: "01001-000", want: "01001000"},
		{name: "spaces", input: "  01001-000  ", want: "01001000"},
		{name: "mixed_separators", input: "01.001/000", want: "01001000"},
		{name: "invalid_character", input: "01a01000", wantErr: true},
		{name: "invalid_length_short", input: "12345", wantErr: true},
		{name: "invalid_length_long", input: "123456789", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCEP(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeCEP() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NormalizeCEP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatCEP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normalized_input", input: "01001000", want: "01001-000"},
		{name: "hyphen_input", input: "01001-000", want: "01001-000"},
		{name: "invalid", input: "abcd", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatCEP(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatCEP() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("FormatCEP() = %q, want %q", got, tt.want)
			}
		})
	}
}
