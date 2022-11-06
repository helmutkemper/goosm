package mongodb

import "goosm/goosm"

type Way struct {
	Id             int64             `bson:"_id"`
	IsPolygon      bool              `bson:"isPolygon"`
	Tag            map[string]string `bson:"tag,omitempty"`
	Loc            GeoJSonLineString `bson:"loc"`
	LocFirst       [2]float64        `bson:"locFirst"`
	LocLast        [2]float64        `bson:"locLast"`
	IdList         []int64           `bson:"idList,omitempty"`
	DistanceTotal  float64           `bson:"distanceTotal"`
	GeoJSonFeature string            `bson:"geoJSonFeature,omitempty"`
}

func (e Way) ToOsmWay() (way goosm.Way) {

	//way.BBox.BottomLeft = e.BBox.BottomLeft.Node()
	//way.BBox.UpperRight = e.BBox.UpperRight.Node()

	way.Id = e.Id
	way.IsPolygon = e.IsPolygon
	way.Tag = e.Tag
	way.Loc = e.Loc.Coordinates
	way.LocFirst = e.LocFirst
	way.LocLast = e.LocLast
	way.DistanceTotal = e.DistanceTotal
	way.GeoJSonFeature = e.GeoJSonFeature
	return
}

func (e *Way) ToDbWay(way *goosm.Way) (dbWay Way) {

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

	e.Id = way.Id
	e.IsPolygon = way.IsPolygon
	e.Tag = way.Tag
	e.Loc.Type = "LineString"
	e.Loc.Coordinates = way.Loc
	e.LocFirst = way.LocFirst
	e.LocLast = way.LocLast
	e.DistanceTotal = way.DistanceTotal
	e.GeoJSonFeature = way.GeoJSonFeature
	return *e
}
