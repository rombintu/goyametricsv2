package storage

import (
	"testing"
)

func Test_memDriver_Save(t *testing.T) {
	type fields struct {
		data      Data
		storepath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "save empty",
			fields: fields{
				storepath: "test.json",
			},
			wantErr: false,
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &memDriver{
				data:      &tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := m.Save(); (err != nil) != tt.wantErr {
				t.Errorf("memDriver.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// func Test_memDriver_Restore(t *testing.T) {
// 	type fields struct {
// 		data      map[string]interface{}
// 		storepath string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		{
// 			name: "restore",
// 			fields: fields{
// 				storepath: "test.json",
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			m := &memDriver{
// 				data:      tt.fields.data,
// 				storepath: tt.fields.storepath,
// 			}
// 			if err := m.Restore(); (err != nil) != tt.wantErr {
// 				t.Errorf("memDriver.Restore() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if len(m.data) != 1 {
// 				t.Error("memDriver.Restore() data not restored")
// 			}
// 			if _, ok := m.data["foo"]; !ok {
// 				t.Error("memDriver.Restore() 'foo' data not restored")
// 			}
// 		})
// 	}
// }
