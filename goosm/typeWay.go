package goosm

import (
	"errors"
	"math"
	"strconv"
)

type Way struct {
	Common `bson:"-"`

	Id             int64             `bson:"_id"`
	IsPolygon      bool              `bson:"isPolygon"`
	Tag            map[string]string `bson:"tag,omitempty"`
	Loc            [][2]float64      `bson:"loc"`
	LocFirst       [2]float64        `bson:"locFirst"`
	LocLast        [2]float64        `bson:"locLast"`
	DistanceTotal  float64           `bson:"distanceTotal"`
	BBox           Box               `bson:"bbox"`
	GeoJSonFeature string            `bson:"geoJSonFeature,omitempty"`
}

func (e *Way) Init() (err error) {

	var longitudeMax = -999.9
	var longitudeMin = 999.9
	var latitudeMax = -999.9
	var latitudeMin = 999.9

	var pointA = Node{}
	var pointB = Node{}

	for k := range e.Loc {
		longitudeMax = math.Max(longitudeMax, e.Loc[k][0])
		longitudeMin = math.Min(longitudeMin, e.Loc[k][0])
		latitudeMax = math.Max(latitudeMax, e.Loc[k][1])
		latitudeMin = math.Min(latitudeMin, e.Loc[k][1])

		if k != 0 {
			pointA.Init(0, e.Loc[k-1][Longitude], e.Loc[k-1][Latitude], nil)
			pointB.Init(0, e.Loc[k][Longitude], e.Loc[k][Latitude], nil)

			e.DistanceTotal += pointA.DistanceBetweenTwoPoints(pointB)
		}
	}

	err = e.makeBBox(longitudeMin, longitudeMax, latitudeMin, latitudeMax)
	if err != nil {
		return
	}

	if len(e.Loc) != 0 {
		e.LocFirst = e.Loc[0]
		e.LocLast = e.Loc[len(e.Loc)-1]
	}

	e.IsPolygon = e.isPolygon()
	e.deleteTagsUnnecessary(&e.Tag)

	return
}

func (e *Way) makeBBox(longitudeMin, longitudeMax, latitudeMin, latitudeMax float64) (err error) {
	err = e.BBox.BottomLeft.SetLngLatDegrees(longitudeMax, latitudeMin)
	if err != nil {
		return
	}
	e.BBox.BottomLeft.MakeGeoJSonFeature()

	err = e.BBox.UpperRight.SetLngLatDegrees(longitudeMin, latitudeMax)
	if err != nil {
		return
	}
	e.BBox.UpperRight.MakeGeoJSonFeature()

	return
}

func (e *Way) isPolygon() (isPolygon bool) {
	var length = len(e.Loc) - 1
	if length < 2 {
		return false
	}

	if e.Loc[0][0] == e.Loc[length][0] && e.Loc[0][1] == e.Loc[length][1] {
		return true
	}

	return false
}

func (e *Way) MakeGeoJSonFeature() (geoJSonStr string) {
	var geoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathWay(strconv.FormatInt(e.Id, 10), e)
	e.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return e.GeoJSonFeature
}

func (e *Way) MakePolygonSurroundingsACW(meters float64) (polygon NewPolygon, err error) {
	if len(e.Loc) < 3 {
		err = errors.New("the way must have a minimum of three points")
		return
	}

	var list = make([][2]float64, 0)

	// contra o relógio
	e.makePolygonSurroundingsRotation(90.0, meters, &list)
	for k := range list {
		polygon.AddLngLatDegrees(list[k][Longitude], list[k][Latitude])
	}

	for k := len(e.Loc) - 1; k >= 0; k-- {
		polygon.AddLngLatDegrees(e.Loc[k][Longitude], e.Loc[k][Latitude])
	}

	err = polygon.Init()
	if err != nil {
		return
	}

	polygon.MakeGeoJSonFeature()
	return
}

func (e *Way) MakePolygonSurroundingsCW(meters float64) (polygon NewPolygon, err error) {
	if len(e.Loc) < 3 {
		err = errors.New("the way must have a minimum of three points")
		return
	}

	var list = make([][2]float64, 0)

	// contra o relógio
	e.makePolygonSurroundingsRotationV2(90.0, meters, &list)
	for k := range list {
		polygon.AddLngLatDegrees(list[k][Longitude], list[k][Latitude])
	}

	for k := 0; k != len(e.Loc); k++ {
		polygon.AddLngLatDegrees(e.Loc[k][Longitude], e.Loc[k][Latitude])
	}

	err = polygon.Init()
	if err != nil {
		return
	}

	polygon.MakeGeoJSonFeature()
	return
}

func (e *Way) makePolygonSurroundingsRotationV2(rotation, meters float64, list *[][2]float64) {
	var i = 0
	var destLoc [2]float64

	// descobre o angulo do ponto externo A e o próximo ponto interno B
	angle := e.directionBetweenTwoPoints(e.Loc[len(e.Loc)-2], e.Loc[len(e.Loc)-1])

	// faz 1/4 de círculo na ponta externa da linha
	for a := 0; a != 11; a++ {
		destLoc = e.destinationPoint(e.Loc[len(e.Loc)-1], meters, angle)
		*list = append(*list, destLoc)

		angle += rotation / 10.0
	}

	// tendo o seguimento A, B, C, onde A e C não são os pontos externos da linha, adiciona os ponto AB e BC com 90˚ a
	// meters de distância do ponto B
	for i = len(e.Loc) - 2; i >= 1; i-- {
		// ponto AB
		angle = e.directionBetweenTwoPoints(e.Loc[i-1], e.Loc[i])
		angle += rotation
		destLoc = e.destinationPoint(e.Loc[i-1], meters, angle)
		*list = append(*list, destLoc)

		// ponto BC
		angle = e.directionBetweenTwoPoints(e.Loc[i], e.Loc[i+1])
		angle += rotation
		destLoc = e.destinationPoint(e.Loc[i+1], meters, angle)
		*list = append(*list, destLoc)
	}

	// descobre o angulo do ponto externo Z e o próximo ponto interno Y
	i = 1
	angle = e.directionBetweenTwoPoints(e.Loc[i], e.Loc[i+1])
	angle += rotation
	destLoc = e.destinationPoint(e.Loc[i-1], meters, angle)
	*list = append(*list, destLoc)

	// faz 1/4 de círculo na ponta externa da linha
	for a := 0; a != 11; a++ {
		destLoc = e.destinationPoint(e.Loc[i-1], meters, angle)
		*list = append(*list, destLoc)

		angle += rotation / 10.0
	}

	return
}

func (e *Way) makePolygonSurroundingsRotation(rotation, meters float64, list *[][2]float64) {
	var i = 0
	var destLoc [2]float64

	// descobre o angulo do ponto externo A e o próximo ponto interno B
	angle := e.directionBetweenTwoPoints(e.Loc[1], e.Loc[0])

	// faz 1/4 de círculo na ponta externa da linha
	for a := 0; a != 11; a++ {
		destLoc = e.destinationPoint(e.Loc[0], meters, angle)
		*list = append(*list, destLoc)

		angle += rotation / 10.0
	}

	// tendo o seguimento A, B, C, onde A e C não são os pontos externos da linha, adiciona os ponto AB e BC com 90˚ a
	// meters de distância do ponto B
	for i = 2; i != len(e.Loc)-1; i++ {
		// ponto AB
		angle = e.directionBetweenTwoPoints(e.Loc[i], e.Loc[i-1])
		angle += rotation
		destLoc = e.destinationPoint(e.Loc[i-1], meters, angle)
		*list = append(*list, destLoc)

		// ponto BC
		angle = e.directionBetweenTwoPoints(e.Loc[i+1], e.Loc[i])
		angle += rotation
		destLoc = e.destinationPoint(e.Loc[i+1], meters, angle)
		*list = append(*list, destLoc)
	}

	// descobre o angulo do ponto externo Z e o próximo ponto interno Y
	i = len(e.Loc) - 2
	angle = e.directionBetweenTwoPoints(e.Loc[i+1], e.Loc[i])
	angle += rotation
	destLoc = e.destinationPoint(e.Loc[i], meters, angle)
	*list = append(*list, destLoc)

	// faz 1/4 de círculo na ponta externa da linha
	for a := 0; a != 11; a++ {
		destLoc = e.destinationPoint(e.Loc[i+1], meters, angle)
		*list = append(*list, destLoc)

		angle += rotation / 10.0
	}

	return
}

// pointSide
//
// English:
//
// Given line L, determine which side of the line point P is
//
// Português:
//
// Dada a linha L, determinar em que lado da linha o ponto P se encontra
//func (e *Way) pointSide(line [][2]float64, point [2]float64) (side float64) {
//	return (line[1][Longitude]-line[0][Longitude])*(point[Latitude]-line[0][Latitude]) - (line[1][Latitude]-line[0][Latitude])*(point[Longitude]-line[0][Longitude])
//}

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
