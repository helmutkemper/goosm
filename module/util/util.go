package util

import (
	"math"
)

func DegreesToRadians(a float64) float64 { return math.Pi * a / 180.0 }

func RadiansToDegrees(a float64) float64 { return 180.0 * a / math.Pi }

func Pythagoras(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2.0) + math.Pow(y2-y1, 2.0))
}

func Round(value float64) float64 {
	var roundOn = 0.5
	var places = 7.0

	var round float64
	pow := math.Pow(10, places)
	digit := pow * value
	_, div := math.Modf(digit)

	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return round / pow
}

func Equal(pointA, pointB [2]float64) (equal bool) {
	if Round(pointA[0]) != Round(pointB[0]) || Round(pointA[1]) != Round(pointB[1]) {
		return false
	}

	return true
}
