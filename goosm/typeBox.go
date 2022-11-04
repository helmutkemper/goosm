package goosm

import (
	"fmt"
	"strconv"
)

type Box struct {
	BottomLeft     Node
	UpperRight     Node
	GeoJSonFeature string `bson:"-"`
}

type BoxList struct {
	List []Box
}

func (e *Box) ToDegreesString() string {
	return fmt.Sprintf("((%1.5f,%1.5f),(%1.5f,%1.5f))%v", e.BottomLeft.Loc[0], e.BottomLeft.Loc[1], e.UpperRight.Loc[0], e.UpperRight.Loc[1], DEGREES)
}

// Get angle as string

func (e *Box) ToRadiansString() string {
	return fmt.Sprintf("((%1.5f,%1.5f),(%1.5f,%1.5f))%v", e.BottomLeft.Rad[0], e.BottomLeft.Rad[1], e.UpperRight.Rad[0], e.UpperRight.Rad[1], RADIANS)
}

func (e *Box) ToGoogleMapString() string {
	return fmt.Sprintf("BottomLeft: %1.5f, %1.5f [ Please, copy and past this value on google maps search ]\nUpperRight: %1.5f, %1.5f [ Please, copy and past this value on google maps search ]", e.BottomLeft.Loc[1], e.BottomLeft.Loc[0], e.UpperRight.Loc[1], e.UpperRight.Loc[0])
}

// Make
//
// English:
//
// Make a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the
// points contained within a rectangular box.
//
// Português:
//
// Monta uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos
// dentro de uma caixa retangular
func (e *Box) Make(pointA Node, meters float64) (err error) {
	var box Box
	box, err = pointA.BoundingBox(meters)

	e.BottomLeft = box.BottomLeft
	e.UpperRight = box.UpperRight

	return
}

// MakeGeoJSonFeature
//
// English:
//
// # Create, archive and return a geoJSon feature
//
// Português:
//
// Cria, arquiva e retorna uma feature geoJSon
func (e *Box) MakeGeoJSonFeature() string {

	// ul   ur
	// bl   br

	var polygon = Polygon{}
	polygon.Tag = make(map[string]string)
	polygon.AddLngLatDegrees(e.UpperRight.Loc[Longitude], e.UpperRight.Loc[Latitude])
	polygon.AddLngLatDegrees(e.BottomLeft.Loc[Longitude], e.UpperRight.Loc[Latitude])
	polygon.AddLngLatDegrees(e.BottomLeft.Loc[Longitude], e.BottomLeft.Loc[Latitude])
	polygon.AddLngLatDegrees(e.UpperRight.Loc[Longitude], e.BottomLeft.Loc[Latitude])

	_ = polygon.Init()

	var geoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygon(strconv.FormatInt(0, 10), &polygon)
	e.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return e.GeoJSonFeature
}
