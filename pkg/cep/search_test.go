package cep

import (
	"testing"
	"time"

	"github.com/jeffotoni/gocep/config"
	"github.com/jeffotoni/gocep/models"
	"github.com/jeffotoni/gocep/service/gocache"
)

// go test -run ^TestSearch$ -v
func TestSearch(t *testing.T) {
	oldCacheEnable := config.CACHE_ENABLE
	config.CACHE_ENABLE = true
	t.Cleanup(func() {
		config.CACHE_ENABLE = oldCacheEnable
	})

	gocache.SetTTL("08226024",
		`{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
		time.Duration(config.TTlCache)*time.Second)
	gocache.SetTTL("01001000", `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`, time.Duration(config.TTlCache)*time.Second)

	type args struct {
		cep string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantCep bool
		wantErr bool
	}{
		{
			name:    "test_search_1",
			args:    args{"08226024"},
			want:    `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
			wantCep: true,
			wantErr: false, // (err != nil) => false
		},
		{
			name:    "test_search_2",
			args:    args{"01001000"},
			want:    `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`,
			wantCep: true,
			wantErr: false,
		},
		{
			name:    "test_search_3",
			args:    args{""},
			want:    `{"cidade":"","uf":"","logradouro":"","bairro":""}`,
			wantCep: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, wecep, err := Search(tt.args.cep)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}

			if ValidCep(wecep) != tt.wantCep {
				t.Errorf("ValidCep() = %v, want %v", ValidCep(wecep), tt.wantCep)
			}
		})
	}
}

// go test -run ^TestValidCep$ -v
func TestValidCep(t *testing.T) {
	type args struct {
		wecep models.WeCep
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_valid_cep_",
			args: args{models.WeCep{Cidade: "São Paulo", Uf: "SP", Logradouro: "Rua Esperança", Bairro: "Cidade Antônio Estevão de Carvalho"}},
			want: true,
		},
		{
			name: "test_valid_cep_",
			args: args{models.WeCep{}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidCep(tt.args.wecep)
			if got != tt.want {
				t.Errorf("ValidCep() = %v, want %v", got, tt.want)
			}
		})
	}
}
