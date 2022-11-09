package goosm

import (
	"errors"
	"fmt"
	"goosm/module/util"
	"math"
	"strconv"
)

// Node
//
// English:
//
// # Create street maps node
//
// Português:
//
// Node do open street maps
type Node struct {
	Common

	// English: Create street maps ID
	// Português: ID do open street maps
	Id int64

	// English: Map ready geolocation array (0:x:longitude,1:y:latitude)
	// Português: Array de localização geográfica pronto para o mapa (0:x:longitude,1:y:latitude)
	Loc [2]float64

	// English: Array Loc converted to radians, used in golang calculations
	// Português: Array Loc convertido para radianos, usado em cálculos golang
	Rad [2]float64

	// Tags do Create Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.

	// English: Tag de descrição do ponto (valores desnecessários são apagados).
	// Português: Point description tag (unnecessary values are deleted).
	Tag map[string]string

	// English: geoJSon feature (GUI).
	// Português: geoJSon feature (GUI).
	GeoJSonFeature string
}

// String
//
// English:
//
// Prints the Node in the form of GeoJSon.
//
// Português:
//
// Imprime o Node na forma de GeoJSon.
func (e Node) String() (geoJson string) {
	if e.GeoJSonFeature == "" {
		e.MakeGeoJSonFeature()
	}

	return fmt.Sprintf("{\"type\":\"FeatureCollection\",\"features\":[%v]}", e.GeoJSonFeature)
}

// Init
//
// English:
//
// Initialize the node. As an example, google maps uses latitude, longitude in the GUI.
//
// Português:
//
// Inicializa o node. Como exemplo, o google maps usa latitude, longitude na interface gráfica.
func (e *Node) Init(id int64, longitude, latitude float64, tags *map[string]string) {
	longitude = util.Round(longitude)
	latitude = util.Round(latitude)

	e.deleteTagsUnnecessary(tags)

	e.Id = id
	e.Loc = [2]float64{longitude, latitude}
	e.Rad = [2]float64{util.DegreesToRadians(longitude), util.DegreesToRadians(latitude)}

	if tags == nil {
		e.Tag = make(map[string]string)
	} else {
		e.Tag = *tags
	}
}

// Set angle value as radians

//func (e *Node) setXYRadians(x, y float64) error {
//	e.Loc = [2]float64{util.RadiansToDegrees(x), util.RadiansToDegrees(y)}
//	e.Rad = [2]float64{x, y}
//
//	return e.checkBounds()
//}

// checkBounds
//
// English:
//
// Determines whether longitude is within the range of -180° to +180° and latitude is within the range of -90° to plus +90°
//
// Português:
//
// Determina se longitude está dentro do alcance de -180º a +180º e latitude dentro do alcance de -90º a mais +90º
func (e *Node) checkBounds() error {
	if e.getLatitudeAsRadians() < MIN_LAT || e.getLatitudeAsRadians() > MAX_LAT {
		return errors.New(fmt.Sprintf("Error: Latitude must be < [math.Pi/2 rad|+90º] and > [-math.Pi/2 rad|-90º]. Value %v\n", e.toRadiansString()))
	}
	if e.getLongitudeAsRadians() < MIN_LON || e.getLongitudeAsRadians() > MAX_LON {
		return errors.New(fmt.Sprintf("Error: Longitude must be < [math.Pi rad|+180º] and > [-math.Pi rad|-180º]. Value %v\n", e.toRadiansString()))
	}

	return nil
}

// Get angle as string

func (e *Node) toRadiansString() string {
	return fmt.Sprintf("(%1.5f,%1.5f)%v", e.Rad[0], e.Rad[1], RADIANS)
}

func (e *Node) getLatitudeAsRadians() float64 { return e.Rad[1] }

func (e *Node) getLongitudeAsRadians() float64 { return e.Rad[0] }

// MakeGeoJSonFeature
//
// English:
//
// # Create, archive and return a geoJSon feature
//
// Português:
//
// Cria, arquiva e retorna uma feature geoJSon
func (e *Node) MakeGeoJSonFeature() string {
	var geoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPoint(strconv.FormatInt(e.Id, 10), e)
	e.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return e.GeoJSonFeature
}

// Set angle value as degrees

//func (e *Node) SetXYDegrees(x, y float64) error {
//	e.Loc = [2]float64{x, y}
//	e.Rad = [2]float64{util.DegreesToRadians(x), util.DegreesToRadians(y)}
//
//	return e.checkBounds()
//}

// Set longitude and latitude as degrees

// SetLngLatDegrees
//
// English:
//
// # Sets longitude and latitude
//
// Português:
//
// Define longitude e latitude
func (e *Node) SetLngLatDegrees(longitude, latitude float64) error {
	e.Loc = [2]float64{longitude, latitude}
	e.Rad = [2]float64{util.DegreesToRadians(longitude), util.DegreesToRadians(latitude)}

	return e.checkBounds()
}

// fixme está estranho...

//func (e *Node) SetLatLngDecimalDrees(latitudeDegrees, latitudePrimes, latitudeSeconds, longitudeDegrees, longitudePrimes, longitudeSeconds int64) {
//	e.Loc = [2]float64{float64(latitudeDegrees) + float64(latitudePrimes)/60.0 + float64(latitudeSeconds)/3600.0, float64(longitudeDegrees) + float64(longitudePrimes)/60.0 + float64(longitudeSeconds)/3600.0}
//	e.Rad = [2]float64{util.DegreesToRadians(float64(latitudeDegrees) + float64(latitudePrimes)/60.0 + float64(latitudeSeconds)/3600.0), util.DegreesToRadians(float64(longitudeDegrees) + float64(longitudePrimes)/60.0 + float64(longitudeSeconds)/3600.0)}
//}

// fixme está estranho...

//func (e *Node) SetLngLatDecimalDrees(longitudeDegrees, longitudePrimes, longitudeSeconds, latitudeDegrees, latitudePrimes, latitudeSeconds int64) {
//	e.Loc = [2]float64{float64(latitudeDegrees) + float64(latitudePrimes)/60.0 + float64(latitudeSeconds)/3600.0, float64(longitudeDegrees) + float64(longitudePrimes)/60.0 + float64(longitudeSeconds)/3600.0}
//	e.Rad = [2]float64{util.DegreesToRadians(float64(latitudeDegrees) + float64(latitudePrimes)/60.0 + float64(latitudeSeconds)/3600.0), util.DegreesToRadians(float64(longitudeDegrees) + float64(longitudePrimes)/60.0 + float64(longitudeSeconds)/3600.0)}
//}

// Set latitude and longitude as radians

//func (e *Node) SetLatLngRadians(latitude, longitude float64) error {
//	e.Loc = [2]float64{util.RadiansToDegrees(longitude), util.RadiansToDegrees(latitude)}
//	e.Rad = [2]float64{longitude, latitude}
//
//	return e.checkBounds()
//}

// Set longitude and latitude as radians

// SetLngLatRadians
//
// English:
//
// # Sets longitude and latitude as radians
//
// Português:
//
// Define longitude e latitude como radianos
func (e *Node) SetLngLatRadians(longitude, latitude float64) error {
	e.Loc = [2]float64{util.RadiansToDegrees(longitude), util.RadiansToDegrees(latitude)}
	e.Rad = [2]float64{longitude, latitude}

	return e.checkBounds()
}

//func (e *Node) SetLatLngRadiansWithoutCheckingFunction(latitude, longitude float64) {
//	e.Loc = [2]float64{util.RadiansToDegrees(longitude), util.RadiansToDegrees(latitude)}
//	e.Rad = [2]float64{longitude, latitude}
//}

func (e *Node) pythagoras(point Node) float64 {
	return math.Sqrt(e.distanceSquared(point))
}

func (e *Node) distanceSquared(point Node) float64 {
	return (math.Abs(point.Loc[0])-math.Abs(e.Loc[0]))*(math.Abs(point.Loc[0])-math.Abs(e.Loc[0])) +
		(math.Abs(point.Loc[1])-math.Abs(e.Loc[1]))*(math.Abs(point.Loc[1])-math.Abs(e.Loc[1]))
}

// DestinationPoint
//
// English:
//
// # Calculate new point at given distance and angle
//
// Português:
//
// Calcular novo ponto em função da distância e do ângulo
func (e *Node) DestinationPoint(meters float64, degrees float64) (newPoint Node, err error) {
	radians := util.DegreesToRadians(degrees)
	earthRadius := EarthRadius(*e)

	y := math.Asin(math.Sin(e.Rad[Latitude])*math.Cos(meters/earthRadius.GetMeters()) +
		math.Cos(e.Rad[Latitude])*math.Sin(meters/earthRadius.GetMeters())*math.Cos(radians))

	x := e.Rad[Longitude] + math.Atan2(math.Sin(radians)*
		math.Sin(meters/earthRadius.GetMeters())*math.Cos(e.Rad[Latitude]),
		math.Cos(meters/earthRadius.GetMeters())-math.Sin(e.Rad[Latitude])*math.Sin(y))

	err = newPoint.SetLngLatRadians(x, y)
	return
}

// DirectionBetweenTwoPoints
//
// English:
//
// Calculate angle between two points.
//
// Português:
//
// Calcula o ângulo entre dois pontos.
func (e *Node) DirectionBetweenTwoPoints(pointB Node) (degrees float64) {
	var radians float64

	y := math.Sin(pointB.getLongitudeAsRadians()-e.getLongitudeAsRadians()) *
		math.Cos(pointB.getLatitudeAsRadians())

	x := math.Cos(e.getLatitudeAsRadians())*math.Sin(pointB.getLatitudeAsRadians()) -
		math.Sin(e.getLatitudeAsRadians())*math.Cos(pointB.getLatitudeAsRadians())*
			math.Cos(pointB.getLongitudeAsRadians()-e.getLongitudeAsRadians())

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

// DistanceBetweenTwoPoints
//
// English:
//
// Calculate distance between two points.
//
// Português:
//
// Calcula a distância entre dois pontos.
func (e *Node) DistanceBetweenTwoPoints(pointB Node) (meters float64) {
	earthRadiusA := EarthRadiusRadLatitude(e.Rad[1])

	meters = math.Acos(math.Sin(e.Rad[1])*math.Sin(pointB.Rad[1])+
		math.Cos(e.Rad[1])*math.Cos(pointB.Rad[1])*
			math.Cos(e.Rad[0]-pointB.Rad[0])) *
		earthRadiusA.GetMeters()

	if math.IsNaN(meters) {
		meters = 0
	}

	return meters
}

// BoundingBox
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
func (e *Node) BoundingBox(meters float64) (box Box, err error) {
	var angle Angle
	angle.SetDegrees(-135.0)
	box.BottomLeft, err = e.DestinationPoint(meters, angle.GetAsDegrees())

	angle.SetDegrees(45.0)
	box.UpperRight, err = e.DestinationPoint(meters, angle.GetAsDegrees())
	return
}
