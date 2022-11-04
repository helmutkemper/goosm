package goosm

import (
	"math"
)

// EarthRadius
//
// English:
//
// # Earth radius at a given latitude, according to the GEOIDAL CONST ellipsoid
//
// Português:
//
// Raio da Terra em uma determinada latitude, de acordo com a constante elipsóide GEOIDAL
func EarthRadius(point Node) Distance {
	return EarthRadiusRadLatitude(point.Rad[1])
}

// EarthRadiusRadLatitude
//
// English:
//
// Earth radius at a given latitude, according to the GEOIDAL CONST ellipsoid
//
//	Input:
//	  latitude: latitude in radians
//
// Português:
//
// Raio da Terra em uma determinada latitude, de acordo com a constante elipsóide GEOIDAL
//
//	Entrada:
//	  latitude: latitude em radianos
func EarthRadiusRadLatitude(latitude float64) Distance {
	var distance Distance

	distance.SetMeters(
		math.Sqrt(
			(math.Pow(math.Pow(GEOIDAL_MAJOR, 2.0)*math.Cos(latitude), 2.0) +
				math.Pow(math.Pow(GEOIDAL_MINOR, 2.0)*math.Sin(latitude), 2.0)) /
				(math.Pow(GEOIDAL_MAJOR*math.Cos(latitude), 2.0) +
					math.Pow(GEOIDAL_MINOR*math.Sin(latitude), 2.0))))

	return distance
}
