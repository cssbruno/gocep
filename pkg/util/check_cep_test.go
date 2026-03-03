package util

import (
	"fmt"
	"testing"
)

// ExampleCheckCEP validates an invalid CEP input.
func ExampleCheckCEP() {
	err := CheckCEP("not-a-cep")
	fmt.Println(err)
	// Output: invalid cep
}

// go test -run ^TestCheckCEP'$ -v
func TestCheckCEP(t *testing.T) {
	type args struct {
		cep string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test_chekcep_", args{"08226021"}, false},
		{"test_chekcep_", args{"01010001"}, false},
		{"test_chekcep_", args{"01010900"}, false},
		{"test_chekcep_", args{"xxxxxxxxx"}, true},
		{"test_chekcep_", args{"1234567"}, true},
		{"test_chekcep_", args{"123456789"}, true},
		{"test_chekcep_", args{"abc12345"}, true},
		{"test_chekcep_", args{"#$%&*^@"}, true},
		{"test_chekcep_", args{""}, true},
		{"test_chekcep_", args{"      "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckCEP(tt.args.cep); (err != nil) != tt.wantErr {
				t.Errorf("CheckCEP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkCheckCEPValid(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := CheckCEP("08226021"); err != nil {
			b.Fatalf("CheckCEP() error = %v", err)
		}
	}
}
