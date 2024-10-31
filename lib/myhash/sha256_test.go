// Package myhash provides utility functions for generating and validating SHA256 HMAC hashes.
// It also includes an Echo middleware for checking the integrity of request bodies using HMAC hashes.
package myhash

import "testing"

func TestToSHA256AndHMAC(t *testing.T) {
	type args struct {
		src []byte
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple_bytes_2_hmac",
			args: args{src: []byte("hello"), key: "secret-key"},
			want: "98e7ffb964bb5a3f902db1fc101a5baa98b6f2cd56858210c9d70f26ac762fc7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSHA256AndHMAC(tt.args.src, tt.args.key); got != tt.want {
				t.Errorf("ToSHA256AndHMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}
