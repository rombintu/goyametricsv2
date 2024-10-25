package internal

import (
	"testing"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

func TestMetrics_setDelta(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	type args struct {
		delta int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "set delta with minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Delta: int64Ptr(-1),
			},
		},
		{
			name: "set delta with plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Delta: int64Ptr(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}
			m.setDelta(tt.args.delta)
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestMetrics_setValue(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	type args struct {
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "set value with minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Value: float64Ptr(-1),
			},
		},
		{
			name: "set value with plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Value: float64Ptr(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}
			m.setValue(tt.args.value)
		})
	}
}

func TestMetrics_SetValueOrDelta(t *testing.T) {
	type fields struct {
		ID    string
		MType string
		Delta *int64
		Value *float64
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set with minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Delta: int64Ptr(1),
				Value: float64Ptr(-1),
			},
			wantErr: true,
		},
		{
			name: "set with plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Delta: int64Ptr(1),
				Value: float64Ptr(1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
			}
			if err := m.SetValueOrDelta(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.SetValueOrDelta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
