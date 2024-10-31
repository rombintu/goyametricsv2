package internal

import (
	"testing"

	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/ptrhelper"
	"github.com/stretchr/testify/assert"
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
			name: "set_delta_with_minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Delta: ptrhelper.Int64Ptr(-1),
			},
		},
		{
			name: "set_delta_with_plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Delta: ptrhelper.Int64Ptr(1),
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
			assert.Equal(t, &tt.args.delta, m.Delta, "Delta should be set correctly")
		})
	}
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
			name: "set_value_with_minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Value: ptrhelper.Float64Ptr(-1),
			},
		},
		{
			name: "se_ value_with_plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Value: ptrhelper.Float64Ptr(1),
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
			assert.Equal(t, &tt.args.value, m.Value, "Value should be set correctly")
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
			name: "set_with_minus",
			fields: fields{
				ID:    "1",
				MType: storage.CounterType,
				Delta: ptrhelper.Int64Ptr(1),
				Value: ptrhelper.Float64Ptr(-1),
			},
			wantErr: true,
		},
		{
			name: "set_with_plus",
			fields: fields{
				ID:    "2",
				MType: storage.CounterType,
				Delta: ptrhelper.Int64Ptr(1),
				Value: ptrhelper.Float64Ptr(1),
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
