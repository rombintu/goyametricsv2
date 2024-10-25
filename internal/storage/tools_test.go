package storage

import (
	"reflect"
	"testing"
)

func Test_counters2Any(t *testing.T) {
	source := make(Counters)
	payload := make(AnyMetrics)
	source["c1"] = 1
	source["c2"] = 2
	payload["c1"] = "1"
	payload["c2"] = "2"
	type args struct {
		source Counters
	}
	tests := []struct {
		name string
		args args
		want AnyMetrics
	}{
		{
			name: "counters2AnyMetrics",
			args: args{source: source},
			want: payload,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := counters2Any(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("counters2Any() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gauges2Any(t *testing.T) {
	source := make(Gauges)
	payload := make(AnyMetrics)
	source["g1"] = 1
	source["g2"] = 2
	payload["g1"] = "1"
	payload["g2"] = "2"
	type args struct {
		source Gauges
	}
	tests := []struct {
		name string
		args args
		want AnyMetrics
	}{
		{
			name: "counters2AnyMetrics",
			args: args{source: source},
			want: payload,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gauges2Any(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gauges2Any() = %v, want %v", got, tt.want)
			}
		})
	}
}
