package myparser

import "strconv"

func Str2Float64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func Str2Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
