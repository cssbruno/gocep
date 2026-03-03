package cep

import (
	"reflect"
	"testing"

	"github.com/cssbruno/gocep/models"
)

func TestParseCEPAddress(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		body    []byte
		want    models.CEPAddress
		wantErr bool
	}{
		{
			name:   "parse cdnapicep",
			source: models.SourceCdnApiCep,
			body:   []byte(`{"status":200,"code":"01001-000","state":"SP","city":"São Paulo","district":"Sé","address":"Praça da Sé - lado ímpar"}`),
			want: models.CEPAddress{
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
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse viacep with padded fields",
			source: models.SourceViaCep,
			body:   []byte(`{"cep":"01001-000","logradouro":"  Praça da Sé  ","localidade":"  São Paulo\t","uf":" sp ","bairro":"\nSé "}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse githubjeffotoni",
			source: models.SourceGitHubJeffotoni,
			body:   []byte(`{"cep":"01001-000","logradouro":"da Sé","bairro":"Sé","uf":"SP","cidade":"São Paulo"}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse postmon",
			source: models.SourcePostmon,
			body:   []byte(`{"bairro":"Sé","cidade":"São Paulo","logradouro":"Praça da Sé","estado":"SP"}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse republica virtual",
			source: models.SourceRepublicaVirtual,
			body:   []byte(`{"uf":"SP","cidade":"São Paulo","bairro":"Sé","logradouro":"da Sé"}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse brasilapi",
			source: models.SourceBrasilAPI,
			body:   []byte(`{"cep":"01001-000","state":"SP","city":"São Paulo","neighborhood":"Sé","street":"Praça da Sé"}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse opencep",
			source: models.SourceOpenCEP,
			body:   []byte(`{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`),
			want: models.CEPAddress{
				City:         "São Paulo",
				StateCode:    "SP",
				Street:       "Praça da Sé",
				Neighborhood: "Sé",
			},
			wantErr: false,
		},
		{
			name:   "parse awesomeapi",
			source: models.SourceAwesomeAPI,
			body:   []byte(`{"cep":"01001000","address":"Praça da Sé","state":"SP","district":"Sé","city":"São Paulo"}`),
			want: models.CEPAddress{
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
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json cdnapicep",
			source:  models.SourceCdnApiCep,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json githubjeffotoni",
			source:  models.SourceGitHubJeffotoni,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json postmon",
			source:  models.SourcePostmon,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json republica virtual",
			source:  models.SourceRepublicaVirtual,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json brasilapi",
			source:  models.SourceBrasilAPI,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json opencep",
			source:  models.SourceOpenCEP,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "invalid json awesomeapi",
			source:  models.SourceAwesomeAPI,
			body:    []byte(`invalid json`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
		{
			name:    "unknown source",
			source:  "unknown",
			body:    []byte(`{}`),
			want:    models.CEPAddress{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCEPAddress(tt.source, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCEPAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCEPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
