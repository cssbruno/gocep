package env

import (
	"os"
	"strconv"
	"testing"
	"time"
)

func TestGetInt(t *testing.T) {
	type args struct {
		name         string
		defaultValue int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test_getint_", args{name: "MONGO_PORT", defaultValue: 10}, 10},
		{"test_getint_", args{name: "MONGO_PORT", defaultValue: 5}, 5},
		{"test_getint_", args{name: "MONGO_PORT", defaultValue: 2}, 2},
		{"test_getint_", args{name: "MONGO_PORT", defaultValue: -100}, -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInt(tt.args.name, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	type args struct {
		name         string
		defaultValue time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"test_getduration_", args{name: "MONGO_TIME",
			defaultValue: time.Duration(time.Second) * 2}, time.Duration(time.Second) * 2},
		{"test_getduration_", args{name: "MONGO_TIME",
			defaultValue: time.Duration(time.Second) * 60}, time.Duration(time.Second) * 60},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDuration(tt.args.name, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	type args struct {
		name         string
		defaultValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test_getstring_", args{name: "MONGO_USER",
			defaultValue: "mongo-user"}, "mongo-user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetString(tt.args.name, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	type args struct {
		name         string
		defaultValue bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test_getstring_", args{name: "MONGO_DB_EXIST",
			defaultValue: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBool(tt.args.name, tt.args.defaultValue); got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInt_WithEnvValues(t *testing.T) {
	const key = "GOCEP_TEST_GET_INT"
	t.Setenv(key, "42")
	if got := GetInt(key, 0); got != 42 {
		t.Fatalf("GetInt() = %d, want 42", got)
	}

	t.Setenv(key, "invalid")
	if got := GetInt(key, 9); got != 9 {
		t.Fatalf("GetInt() with invalid env = %d, want 9", got)
	}
}

func TestGetInt64_WithEnvValues(t *testing.T) {
	const key = "GOCEP_TEST_GET_INT64"
	t.Setenv(key, "77")
	if got := GetInt64(key, 0); got != 77 {
		t.Fatalf("GetInt64() = %d, want 77", got)
	}

	t.Setenv(key, "invalid")
	if got := GetInt64(key, 11); got != 11 {
		t.Fatalf("GetInt64() with invalid env = %d, want 11", got)
	}

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv() error = %v", err)
	}
	if got := GetInt64(key, 13); got != 13 {
		t.Fatalf("GetInt64() with unset env = %d, want 13", got)
	}
}

func TestGetInt64_LargeAndOverflowValues(t *testing.T) {
	const key = "GOCEP_TEST_GET_INT64_MAX"
	const maxInt64 = int64(9223372036854775807)

	t.Setenv(key, strconv.FormatInt(maxInt64, 10))
	if got := GetInt64(key, 0); got != maxInt64 {
		t.Fatalf("GetInt64() = %d, want %d", got, maxInt64)
	}

	t.Setenv(key, "9223372036854775808")
	if got := GetInt64(key, 99); got != 99 {
		t.Fatalf("GetInt64() overflow fallback = %d, want 99", got)
	}
}

func TestGetDuration_WithEnvValues(t *testing.T) {
	const key = "GOCEP_TEST_GET_DURATION"
	t.Setenv(key, "1500")
	if got := GetDuration(key, 0); got != 1500*time.Millisecond {
		t.Fatalf("GetDuration() = %s, want 1500ms", got)
	}

	t.Setenv(key, "invalid")
	if got := GetDuration(key, 2*time.Second); got != 2*time.Second {
		t.Fatalf("GetDuration() with invalid env = %s, want 2s", got)
	}
}

func TestGetString_WithSetAndUnsetEnv(t *testing.T) {
	const key = "GOCEP_TEST_GET_STRING"
	t.Setenv(key, "value")
	if got := GetString(key, "default"); got != "value" {
		t.Fatalf("GetString() = %q, want %q", got, "value")
	}

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv() error = %v", err)
	}
	if got := GetString(key, "default"); got != "default" {
		t.Fatalf("GetString() with unset env = %q, want %q", got, "default")
	}
}

func TestGetBool_WithEnvValues(t *testing.T) {
	const key = "GOCEP_TEST_GET_BOOL"
	t.Setenv(key, "false")
	if got := GetBool(key, true); got != false {
		t.Fatalf("GetBool() = %v, want false", got)
	}

	t.Setenv(key, "not-bool")
	if got := GetBool(key, true); got != true {
		t.Fatalf("GetBool() with invalid env = %v, want true", got)
	}
}
