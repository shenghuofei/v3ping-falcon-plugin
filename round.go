package main

import "math"

//小数精度处理
func Round(f float64, n int) float64 {
	x := 0.5
	if f < 0 {
		x = -0.5
	}
	pow10_n := math.Pow10(n)
	return math.Trunc((f+x/pow10_n)*pow10_n) / pow10_n
}
