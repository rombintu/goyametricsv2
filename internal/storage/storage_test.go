// Package storage Storage
package storage

import (
	"testing"
)

func TestNewStorage(t *testing.T) {
	type args struct {
		storageType string
		storepath   string
	}
	tests := []struct {
		name string
		args args
		want Storage
	}{
		{
			name: "new_storage_mem",
			args: args{
				storageType: "mem",
				storepath:   "",
			},
			want: &tmpDriver{},
		},
		{
			name: "new_storage_pgx",
			args: args{
				storageType: "pgx",
				storepath:   "",
			},
			want: &pgxDriver{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStorage(tt.args.storageType, tt.args.storepath)
			switch got.(type) {
			case *pgxDriver:
				t.Log("pgxdriver init")
			case *tmpDriver:
				t.Log("tmpdriver init")
			default:
				t.Error("driver unknown")
			}

		})
	}
}
