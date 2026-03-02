package cep

import (
	"reflect"
	"testing"

	"github.com/cssbruno/gocep/models"
)

func TestParseWeCep(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		body    []byte
		want    models.WeCep
		wantErr bool
	}{
		{
			name:   "parse cdnapicep",
			source: models.SourceCdnApiCep,
			body:   []byte(`{"status":200,"code":"01001-000","state":"SP","city":"São Paulo","district":"Sé","address":"Praça da Sé - lado ímpar"}`),
			want: models.WeCep{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé - lado ímpar",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse viacep",
			source: models.SourceViaCep,
			body:   []byte(`{"cep":"01001-000","logradouro":"Praça da Sé","localidade":"São Paulo","uf":"SP","bairro":"Sé"}`),
			want: models.WeCep{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			source:  models.SourceViaCep,
			body:    []byte(`invalid json`),
			want:    models.WeCep{},
			wantErr: true,
		},
		{
			name:    "unknown source",
			source:  "unknown",
			body:    []byte(`{}`),
			want:    models.WeCep{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWeCep(tt.source, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWeCep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseWeCep() = %v, want %v", got, tt.want)
			}
		})
	}
}
