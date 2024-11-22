package mycrypt

import (
	"os"
	"path"
	"testing"
)

func TestValidPrivateKey(t *testing.T) {
	tempfile, err := os.CreateTemp("/tmp", "tempfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempfile.Name())
	if err := GenPrivKeyAndCertPEM(tempfile.Name()); err != nil {
		t.Error(err)
	}
	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "invalid",
			args: args{filePath: "invalid_private.key"},
			want: false,
		},
		{
			name: "valid",
			args: args{filePath: tempfile.Name()},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPrivateKey(tt.args.filePath); got != tt.want {
				t.Errorf("ValidPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenPrivKeyAndCertPEM(t *testing.T) {
	tmpDir := os.TempDir()
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "gen_files",
			args:    args{filePath: path.Join(tmpDir, "master")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenPrivKeyAndCertPEM(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("GenPrivKeyAndCertPEM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
