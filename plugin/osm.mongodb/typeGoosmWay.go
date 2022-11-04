package mongodbosm

import (
	"goosm/goosm"
)

type GeoJSonPoint struct {
	Type        string     `bson:"type"`
	Coordinates [2]float64 `bson:"coordinates"`
}

type GeoJSonLineString struct {
	Type        string       `bson:"type"`
	Coordinates [][2]float64 `bson:"coordinates"`
}

type Distance struct {
	Meters float64
}

type Angle struct {
	Degrees float64 `bson:"degrees"`
	Radians float64 `bson:"radians"`
}

type Box struct {
	BottomLeft Node `bson:"bottomleft"`
	UpperRight Node `bson:"upperright"`
}

type Node struct {
	Id int64 `bson:"_id"`
	// Array de localização geográfica.
	// [0:x:longitude,1:y:latitude]
	// Este campo deve obrigatoriamente ser um array devido a indexação do MongoDB
	Loc GeoJSonPoint `bson:"loc"`
	//Rad [2]float64   `bson:"rad"`

	// Tags do Open Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag map[string]string `bson:"tag,omitempty"`

	// Node usado apenas para o parser do arquivo
	GeoJSonFeature string `bson:"geoJSonFeature,omitempty"`
}

func (e Node) Node() (node goosm.Node) {
	node.Id = e.Id
	node.Tag = e.Tag
	//node.Rad = e.Rad
	node.Loc = e.Loc.Coordinates
	node.GeoJSonFeature = e.GeoJSonFeature
	return
}

type Way struct {
	Id        int64             `bson:"_id"`
	IsPolygon bool              `bson:"isPolygon"`
	Tag       map[string]string `bson:"tag,omitempty"`
	Loc       GeoJSonLineString `bson:"loc"`
	LocFirst  [2]float64        `bson:"locFirst"`
	LocLast   [2]float64        `bson:"locLast"`
	//Rad           [][2]float64      `bson:"-"` //`bson:"rad"`
	IdList []int64 `bson:"idList,omitempty"`
	//Distance      []Distance        `bson:"-"` //`bson:"distance"`
	DistanceTotal float64 `bson:"distanceTotal"`
	//Angle         []Angle           `bson:"-"` //`bson:"angle"`

	// en: boundary box in degrees
	// pt: caixa de perímetro em graus decimais
	//BBox           Box    `bson:"bbox"`
	GeoJSonFeature string `bson:"geoJSonFeature,omitempty"`
	//
	//SurroundingPreset []float64 `bson:"SurroundingPreset" json:"-"`
}

func (e Way) Way() (way goosm.Way) {

	//way.Distance = make([]goosm.Distance, len(e.Distance))
	//for k := range e.Distance {
	//	way.Distance[k].Meters = e.Distance[k].Meters
	//}

	//way.Angle = make([]goosm.Angle, len(e.Angle))
	//for k := range e.Angle {
	//	way.Angle[k].Degrees = e.Angle[k].Degrees
	//	way.Angle[k].Radians = e.Angle[k].Radians
	//}

	//way.BBox.BottomLeft = e.BBox.BottomLeft.Node()
	//way.BBox.UpperRight = e.BBox.UpperRight.Node()

	way.Id = e.Id
	way.IsPolygon = e.IsPolygon
	way.Tag = e.Tag
	way.Loc = e.Loc.Coordinates
	way.LocFirst = e.LocFirst
	way.LocLast = e.LocLast
	//way.Rad = e.Rad
	//way.IdList = e.IdList
	way.DistanceTotal = e.DistanceTotal
	way.GeoJSonFeature = e.GeoJSonFeature
	return
}

func (e Way) ToDbWay(way goosm.Way) (dbWay Way) {

	//dbWay.Distance = make([]Distance, len(way.Distance))
	//for k := range way.Distance {
	//	dbWay.Distance[k].Meters = way.Distance[k].Meters
	//}

	//dbWay.Angle = make([]Angle, len(way.Angle))
	//for k := range way.Angle {
	//	dbWay.Angle[k].Degrees = way.Angle[k].Degrees
	//	dbWay.Angle[k].Radians = way.Angle[k].Radians
	//}

	//way.BBox.BottomLeft.MakeGeoJSonFeature()
	//dbWay.BBox.BottomLeft.Rad = way.BBox.BottomLeft.Rad
	//dbWay.BBox.BottomLeft.Loc.Type = "Point"
	//dbWay.BBox.BottomLeft.Loc.Coordinates = way.BBox.BottomLeft.Loc
	//dbWay.BBox.BottomLeft.Tag = way.BBox.BottomLeft.Tag
	//dbWay.BBox.BottomLeft.Id = way.Id
	//dbWay.BBox.BottomLeft.GeoJSonFeature = way.BBox.BottomLeft.GeoJSonFeature

	//way.BBox.UpperRight.MakeGeoJSonFeature()
	//dbWay.BBox.UpperRight.Rad = way.BBox.UpperRight.Rad
	//dbWay.BBox.UpperRight.Loc.Type = "Point"
	//dbWay.BBox.UpperRight.Loc.Coordinates = way.BBox.UpperRight.Loc
	//dbWay.BBox.UpperRight.Tag = way.BBox.UpperRight.Tag
	//dbWay.BBox.UpperRight.Id = way.Id
	//dbWay.BBox.UpperRight.GeoJSonFeature = way.BBox.UpperRight.GeoJSonFeature

	dbWay.Id = way.Id
	dbWay.IsPolygon = way.IsPolygon
	dbWay.Tag = way.Tag
	dbWay.Loc.Type = "LineString"
	dbWay.Loc.Coordinates = way.Loc
	dbWay.LocFirst = way.LocFirst
	dbWay.LocLast = way.LocLast
	//dbWay.Rad = way.Rad
	//dbWay.IdList = way.IdList
	dbWay.DistanceTotal = way.DistanceTotal
	dbWay.GeoJSonFeature = way.GeoJSonFeature
	return
}
