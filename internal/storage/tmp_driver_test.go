package storage

import (
	"reflect"
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

func Test_tmpDriver_GetAll(t *testing.T) {
	type fields struct {
		data      Data
		storepath string
	}
	tests := []struct {
		name   string
		fields fields
		want   Data
	}{
		{
			name: "get_all_nill",
			fields: fields{
				storepath: "test.json",
			},
			want: Data{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      &tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if got := d.GetAll(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tmpDriver.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tmpDriver_getGauge(t *testing.T) {
	data := &Data{
		Gauges: make(Gauges),
	}
	data.Gauges["g1"] = 1
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
		want1  bool
	}{
		{
			name: "get_gauge_1",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:  args{key: "g1"},
			want:  1,
			want1: true,
		},
		{
			name: "get_gauge_nil",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:  args{key: "g2"},
			want:  0,
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			got, got1 := d.getGauge(tt.args.key)
			if got != tt.want {
				t.Errorf("tmpDriver.getGauge() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("tmpDriver.getGauge() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_tmpDriver_getCounter(t *testing.T) {
	data := &Data{
		Counters: make(Counters),
	}
	data.Counters["c1"] = 1
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  bool
	}{
		{
			name: "get_counter_1",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:  args{key: "c1"},
			want:  1,
			want1: true,
		},
		{
			name: "get_counter_nil",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:  args{key: "c2"},
			want:  0,
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			got, got1 := d.getCounter(tt.args.key)
			if got != tt.want {
				t.Errorf("tmpDriver.getCounter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("tmpDriver.getCounter() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_tmpDriver_Get(t *testing.T) {
	data := &Data{
		Counters: make(Counters),
		Gauges:   make(Gauges),
	}
	data.Counters["c1"] = 1
	data.Gauges["g1"] = 1
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		mtype string
		mname string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get_c1_1",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:    args{mtype: CounterType, mname: "c1"},
			want:    "1",
			wantErr: false,
		},
		{
			name: "get_c2_nil",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args:    args{mtype: CounterType, mname: "c2"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			got, err := d.Get(tt.args.mtype, tt.args.mname)
			if (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("tmpDriver.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tmpDriver_Close(t *testing.T) {
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
			name: "close",
			fields: fields{
				storepath: "test.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      &tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tmpDriver_Ping(t *testing.T) {
	type fields struct {
		data      *Data
		storepath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ping",
			fields: fields{
				storepath: "test.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := d.Ping(); (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tmpDriver_updateGauge(t *testing.T) {
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "update_gauge_g1",
			fields: fields{
				storepath: "test.json",
				data: &Data{
					Counters: make(Counters),
					Gauges:   make(Gauges),
				},
			},
			args: args{key: "g1", value: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			d.updateGauge(tt.args.key, tt.args.value)
		})
	}
}

func Test_tmpDriver_updateCounter(t *testing.T) {
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "update_counter_c1",
			fields: fields{
				storepath: "test.json",
				data: &Data{
					Counters: make(Counters),
					Gauges:   make(Gauges),
				},
			},
			args: args{key: "c1", value: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			d.updateCounter(tt.args.key, tt.args.value)
		})
	}
}

func Test_tmpDriver_UpdateAll(t *testing.T) {
	data := &Data{
		Counters: make(Counters),
		Gauges:   make(Gauges),
	}
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		data Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update_data",
			fields: fields{
				storepath: "test.json",
				data:      data,
			},
			args: args{data: *data},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := d.UpdateAll(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.UpdateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tmpDriver_Update(t *testing.T) {
	data := &Data{
		Counters: make(Counters),
		Gauges:   make(Gauges),
	}
	type fields struct {
		data      *Data
		storepath string
	}
	type args struct {
		mtype  string
		mname  string
		mvalue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update_some_value",
			fields: fields{
				data:      data,
				storepath: "test.json",
			},
			args: args{
				mtype:  CounterType,
				mname:  "testing1",
				mvalue: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := d.Update(tt.args.mtype, tt.args.mname, tt.args.mvalue); (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tmpDriver_Open(t *testing.T) {
	cmap := make(Counters)
	gmap := make(Gauges)
	data := Data{
		Counters: cmap,
		Gauges:   gmap,
	}
	type fields struct {
		data      *Data
		storepath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "open_tmp_driver",
			fields: fields{
				data:      &data,
				storepath: "store-test.json",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &tmpDriver{
				data:      tt.fields.data,
				storepath: tt.fields.storepath,
			}
			if err := d.Open(); (err != nil) != tt.wantErr {
				t.Errorf("tmpDriver.Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTmpDriver(t *testing.T) {
	type args struct {
		storepath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create_new_tmp_driver",
			args: args{storepath: "mem"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTmpDriver(tt.args.storepath)
			if got.storepath != tt.args.storepath {
				t.Error("error create new driver tmp")
			}
		})
	}
}
