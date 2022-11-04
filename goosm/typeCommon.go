package goosm

import (
	"goosm/module/util"
	"math"
)

type Common struct {
}

func (e *Common) deleteTagsUnnecessary(tag *map[string]string) {
	if tag == nil {
		return
	}

	delete(*tag, "source")
	delete(*tag, "Source")
	delete(*tag, "history")
	delete(*tag, "converted_by")
	delete(*tag, "created_by")
	delete(*tag, "wikipedia")
	delete(*tag, "wikidata")
}

// EarthRadius
//
// English:
//
// # Earth radius at a given latitude, according to the GEOIDAL CONST ellipsoid
//
// Português:
//
// Raio da Terra em uma determinada latitude, de acordo com a constante elipsóide GEOIDAL
func (e *Common) earthRadius(point [2]float64) (meters float64) {
	return e.earthRadiusRadLatitude(point[Latitude])
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
func (e *Common) earthRadiusRadLatitude(latitude float64) (meters float64) {
	return math.Sqrt(
		(math.Pow(math.Pow(GEOIDAL_MAJOR, 2.0)*math.Cos(latitude), 2.0) +
			math.Pow(math.Pow(GEOIDAL_MINOR, 2.0)*math.Sin(latitude), 2.0)) /
			(math.Pow(GEOIDAL_MAJOR*math.Cos(latitude), 2.0) +
				math.Pow(GEOIDAL_MINOR*math.Sin(latitude), 2.0)))
}

func (e *Common) pythagoras(pointA, pointB [2]float64) float64 {
	return math.Sqrt(e.distanceSquared(pointA, pointB))
}

func (e *Common) distanceSquared(pointA, pointB [2]float64) float64 {
	return (math.Abs(pointB[Longitude])-math.Abs(pointA[Longitude]))*(math.Abs(pointB[Longitude])-math.Abs(pointA[Longitude])) +
		(math.Abs(pointB[Latitude])-math.Abs(pointA[Latitude]))*(math.Abs(pointB[Latitude])-math.Abs(pointA[Latitude]))
}

// destinationPoint
//
// English:
//
// # Calculate new point at given distance and angle
//
// Português:
//
// Calcular novo ponto em função da distância e do ângulo
func (e *Common) destinationPoint(pointA [2]float64, meters float64, degrees float64) (newPoint [2]float64) {
	pointA[Longitude] = util.DegreesToRadians(pointA[Longitude])
	pointA[Latitude] = util.DegreesToRadians(pointA[Latitude])

	var radians = util.DegreesToRadians(degrees)
	var earthRadius = e.earthRadius(pointA)

	latitude := math.Asin(math.Sin(pointA[Latitude])*math.Cos(meters/earthRadius) +
		math.Cos(pointA[Latitude])*math.Sin(meters/earthRadius)*math.Cos(radians))

	longitude := pointA[Longitude] + math.Atan2(math.Sin(radians)*
		math.Sin(meters/earthRadius)*math.Cos(pointA[Latitude]),
		math.Cos(meters/earthRadius)-math.Sin(pointA[Latitude])*math.Sin(latitude))

	newPoint = [2]float64{e.Round(util.RadiansToDegrees(longitude)), e.Round(util.RadiansToDegrees(latitude))}
	return
}

// directionBetweenTwoPoints
//
// English:
//
// Calculate an angle between two points.
//
// Português:
//
// Calcula o ângulo entre dois pontos.
func (e *Common) directionBetweenTwoPoints(pointA, pointB [2]float64) (degrees float64) {
	pointA[Longitude] = util.DegreesToRadians(pointA[Longitude])
	pointA[Latitude] = util.DegreesToRadians(pointA[Latitude])

	pointB[Longitude] = util.DegreesToRadians(pointB[Longitude])
	pointB[Latitude] = util.DegreesToRadians(pointB[Latitude])

	var radians float64

	y := math.Sin(pointB[Longitude]-pointA[Longitude]) *
		math.Cos(pointB[Latitude])

	x := math.Cos(pointA[Latitude])*math.Sin(pointB[Latitude]) -
		math.Sin(pointA[Latitude])*math.Cos(pointB[Latitude])*
			math.Cos(pointB[Longitude]-pointA[Longitude])

	if y > 0.0 {
		if x > 0.0 {
			radians = math.Atan(y / x)
		}
		if x < 0.0 {
			radians = math.Pi - math.Atan(-y/x)
		}
		if x == 0.0 {
			radians = math.Pi / 2.0
		}
	}
	if y < 0.0 {
		if x > 0.0 {
			radians = -math.Atan(-y/x) + 2.0*math.Pi
		}
		if x < 0.0 {
			radians = math.Atan(y/x) - math.Pi + 2.0*math.Pi
		}
		if x == 0.0 {
			radians = math.Pi * 3.0 / 2.0
		}
	}
	if y == 0.0 {
		if x > 0.0 {
			radians = 0.0
		}
		if x < 0.0 {
			radians = math.Pi
		}
		if x == 0.0 {
			radians = 0.0
		}
	}

	return util.RadiansToDegrees(radians)
}

// distanceBetweenTwoPoints
//
// English:
//
// Calculate distance between two points.
//
// Português:
//
// Calcula a distância entre dois pontos.
func (e *Common) distanceBetweenTwoPoints(pointA, pointB [2]float64) (meters float64) {
	pointA[Longitude] = util.DegreesToRadians(pointA[Longitude])
	pointA[Latitude] = util.DegreesToRadians(pointA[Latitude])

	pointB[Longitude] = util.DegreesToRadians(pointB[Longitude])
	pointB[Latitude] = util.DegreesToRadians(pointB[Latitude])

	var earthRadius = e.earthRadiusRadLatitude(pointA[Latitude])

	meters = math.Acos(math.Sin(pointA[Latitude])*math.Sin(pointB[Latitude])+
		math.Cos(pointA[Latitude])*math.Cos(pointB[Latitude])*
			math.Cos(pointA[Longitude]-pointB[Longitude])) * earthRadius

	if math.IsNaN(meters) {
		meters = 0
	}

	return meters
}

// boundingBox
//
// English:
//
// # Bounding box surrounding the point at given coordinates
//
// The box is formed by two dots, upper right and lower left.
//
// Português:
//
// # Caixa delimitadora em torno do ponto nas coordenadas dadas
//
// A caixa é formada por dois pontos, superior direito e inferior esquerdo.
func (e *Common) boundingBox(pointA [2]float64, meters float64) (box Box) {
	var pointBottomLeft [2]float64
	var pointUpperRight [2]float64
	var angle float64
	angle = -135.0
	pointBottomLeft = e.destinationPoint(pointA, meters, angle)
	box.BottomLeft.Init(0, pointBottomLeft[0], pointBottomLeft[1], nil)

	angle = 45.0
	pointUpperRight = e.destinationPoint(pointA, meters, angle)
	box.UpperRight.Init(0, pointUpperRight[0], pointUpperRight[1], nil)
	return
}

func (e *Common) Round(value float64) float64 {
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

func (e *Common) Equal(pointA, pointB [2]float64) (equal bool) {
	if e.Round(pointA[0]) != e.Round(pointB[0]) || e.Round(pointA[1]) != e.Round(pointB[1]) {
		return false
	}

	return true
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
