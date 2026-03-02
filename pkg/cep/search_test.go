package cep

import (
	"testing"
	"time"

	"github.com/cssbruno/gocep/config"
	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/service/gocache"
)

// go test -run ^TestSearch$ -v
func TestSearch(t *testing.T) {
	oldCacheEnable := config.CacheEnabled
	config.CacheEnabled = true
	t.Cleanup(func() {
		config.CacheEnabled = oldCacheEnable
	})

	gocache.SetTTL("08226024",
		`{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`,
		time.Duration(config.TTLCache)*time.Second)
	gocache.SetTTL("01001000", `{"cidade":"São Paulo","uf":"SP","logradouro":"da Sé","bairro":"Sé"}`, time.Duration(config.TTLCache)*time.Second)

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

			if ValidCEP(wecep) != tt.wantCep {
				t.Errorf("ValidCEP() = %v, want %v", ValidCEP(wecep), tt.wantCep)
			}
		})
	}
}

// go test -run ^TestValidCEP$ -v
func TestValidCEP(t *testing.T) {
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
			args: args{models.WeCep{City: "São Paulo", StateCode: "SP", Street: "Rua Esperança", Neighborhood: "Cidade Antônio Estevão de Carvalho"}},
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
			got := ValidCEP(tt.args.wecep)
			if got != tt.want {
				t.Errorf("ValidCEP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSearchCacheHit(b *testing.B) {
	oldCacheEnable := config.CacheEnabled
	config.CacheEnabled = true
	defer func() {
		config.CacheEnabled = oldCacheEnable
	}()

	cep := "08226024"
	payload := `{"cidade":"São Paulo","uf":"SP","logradouro":"Rua Esperança","bairro":"Cidade Antônio Estevão de Carvalho"}`
	wecep := models.WeCep{
		City:         "São Paulo",
		StateCode:    "SP",
		Street:       "Rua Esperança",
		Neighborhood: "Cidade Antônio Estevão de Carvalho",
	}
	gocache.SetAnyTTL(cep, cachedResult{JSON: payload, WeCep: wecep}, time.Duration(config.TTLCache)*time.Second)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Search(cep)
		if err != nil {
			b.Fatalf("Search() error = %v", err)
		}
	}
}
