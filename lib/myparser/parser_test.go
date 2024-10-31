package myparser

import (
	"testing"
)

func TestStr2Float64(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:    "simple_toString",
			args:    args{"1.0"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "simple_error_toString",
			args:    args{"1-1"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Str2Float64(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Str2Float64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Str2Float64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStr2Int64(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "simple_toString",
			args:    args{"1"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "simple_error_toString",
			args:    args{"1-1"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Str2Int64(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Str2Int64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Str2Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}
