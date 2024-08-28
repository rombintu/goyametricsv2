package storage

import (
	"testing"
)

func Test_tmpDriver_Save(t *testing.T) {
	type fields struct {
		data      Data
		storepath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "saveEmpty",
			fields: fields{
				storepath: "test.json",
			},
		},
		{
			name: "save",
			fields: fields{
				data: Data{
					Gauges: map[string]float64{
						"gauge1": 10.0,
						"gauge2": 20.0,
					},
					Counters: map[string]int64{
						"counter1": 1,
						"counter2": 2,
					},
				},
				storepath: "test.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &tmpDriver{
				data:      &tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := m.Save(); err != nil {
				t.Errorf("tmpDriver.Save() error = %v", err)
			}
		})
	}
}

func Test_tmpDriver_Restore(t *testing.T) {
	type fields struct {
		data      Data
		storepath string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "restore",
			fields: fields{
				storepath: "test.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &tmpDriver{
				data:      &tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := m.Restore(); err != nil {
				t.Errorf("tmpDriver.Restore() error = %v", err)
			}
			if len(m.data.Counters) != 2 {
				t.Error("tmpDriver.Restore() len Counters not 2")
			}
			if _, ok := m.getCounter("counter2"); !ok {
				t.Error("tmpDriver.Restore() 'counter2' data not restored")
			}
		})
	}
}
