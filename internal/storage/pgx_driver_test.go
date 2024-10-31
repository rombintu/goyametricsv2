package storage

import (
	"errors"
	"net"
	"testing"
)

const testCredsURL = "host=localhost user=admin password=admin dbname=metrics sslmode=disable"

func Test_pgxDriver_Ping(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "ping_pgx",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			defer db.Close()
			if err := db.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_pgxDriver_Open(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "pgx_open",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					t.Skipf("Skipping test due to network timeout error: %v", err)
				} else if errors.Is(err, &net.OpError{}) {
					t.Skipf("Skipping test due to network operation error: %v", err)
				}
			} else if (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer db.Close()
		})
	}
}

func Test_pgxDriver_Close(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "pgx_close",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			if err := db.Close(); (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_pgxDriver_Save(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "pgx_save_skip",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			defer db.Close()
			if err := db.Save(); (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_pgxDriver_Restore(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "pgx_restore_skip",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			defer db.Close()
			if err := db.Restore(); (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.Restore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_pgxDriver_GetAll(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "pgx_getall_nill_data",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			defer db.Close()
			if got := db.GetAll(); len(got.Counters) == 0 {
				t.Errorf("pgxDriver.GetAll() = %v, want %v", got, tt.want)
			}
			if got := db.GetAll(); len(got.Gauges) == 0 {
				t.Errorf("pgxDriver.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pgxDriver_createTables(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "pgx_create_table",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewPgxDriver(testCredsURL)
			if err := db.Open(); err != nil {
				t.Skipf("Skipping test due to database connection error: %v", err)
			}
			defer db.Close()
			if err := db.createTables(); (err != nil) != tt.wantErr {
				t.Errorf("pgxDriver.createTables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
