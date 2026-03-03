package util

import (
	"testing"

	"github.com/cssbruno/gocep/models"
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name string
		in   models.CEPAddress
		want models.CEPAddress
	}{
		{
			name: "trim city street neighborhood and uppercase state code",
			in: models.CEPAddress{
				CEP:          "01001-000",
				City:         "  São Paulo\t",
				StateCode:    "  sp ",
				Street:       "  Praça da Sé  ",
				Neighborhood: "\n Sé ",
			},
			want: models.CEPAddress{
				CEP:          "01001-000",
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
		},
		{
			name: "blank values remain blank after trim",
			in: models.CEPAddress{
				City:         " ",
				StateCode:    "   ",
				Street:       "\t",
				Neighborhood: "\n",
			},
			want: models.CEPAddress{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeAddress(tt.in)
			if got != tt.want {
				t.Fatalf("NormalizeAddress() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
