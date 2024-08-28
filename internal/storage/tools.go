package storage

import (
	"strconv"
)

// Lib tools
func counters2Any(source Counters) AnyMetrics {
	newMap := make(AnyMetrics)
	for k, v := range source {
		newMap[k] = strconv.FormatInt(v, 10)
	}
	return newMap
}

func gauges2Any(source Gauges) AnyMetrics {
	newMap := make(AnyMetrics)
	for k, v := range source {
		newMap[k] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	return newMap
}
