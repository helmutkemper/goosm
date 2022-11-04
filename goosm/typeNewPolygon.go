package goosm

import (
	"math"
	"strconv"
)

type NewPolygon struct {
	Common

	Id             int64
	Tag            map[string]string
	Loc            [][2]float64
	Length         int
	BBox           Box
	GeoJSonFeature string
	DistanceTotal  float64
}

func (e *NewPolygon) AddLngLatDegrees(longitude, latitude float64) {
	if e.Loc == nil {
		e.Loc = make([][2]float64, 0)
	}

	e.Loc = append(e.Loc, [2]float64{longitude, latitude})
}

func (e *NewPolygon) AddNode(node Node) {
	if e.Loc == nil {
		e.Loc = make([][2]float64, 0)
	}

	e.Loc = append(e.Loc, [2]float64{node.Loc[Longitude], node.Loc[Latitude]})
}

func (e *NewPolygon) Init() (err error) {

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

	e.deleteTagsUnnecessary(&e.Tag)
	return
}

func (e *NewPolygon) makeBBox(longitudeMin, longitudeMax, latitudeMin, latitudeMax float64) (err error) {
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

func (e *NewPolygon) MakeGeoJSonFeature() string {

	// fixme: fazer
	//if el.Id == 0 {
	//	el.Id = util.AutoId.Get(el.DbCollectionName)
	//}

	var geoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathNewPolygon(strconv.FormatInt(e.Id, 10), e)
	e.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return e.GeoJSonFeature
}

func (e *NewPolygon) pointInPolygonDirectionOnLine(line [2][2]float64, loc [2]float64) (onLine bool) {
	// Check whether p is on the line or not
	if loc[Longitude] <= math.Max(line[0][Longitude], line[1][Longitude]) &&
		loc[Longitude] <= math.Min(line[0][Longitude], line[1][Longitude]) &&
		(loc[Latitude] <= math.Max(line[0][Latitude], line[1][Latitude]) &&
			loc[Latitude] <= math.Min(line[0][Latitude], line[1][Latitude])) {
		return true
	}

	return false
}

func (e *NewPolygon) pointInPolygonDirection(a, b, c [2]float64) (direction int) {
	d := (b[Latitude]-a[Latitude])*(c[Longitude]-b[Longitude]) - (b[Longitude]-a[Longitude])*(c[Latitude]-b[Latitude])
	if d == 0 {
		// Colinear
		return 0
	} else if d < 0 {
		// Anti-clockwise direction
		return 2
	}

	// Clockwise direction
	return 1
}

func (e *NewPolygon) pointInPolygonIsIntersect(l1, l2 [2][2]float64) (isIntersect bool) {
	// Four direction for two lines and points of other line
	dir1 := e.pointInPolygonDirection(l1[0], l1[1], l2[0])
	dir2 := e.pointInPolygonDirection(l1[0], l1[1], l2[1])
	dir3 := e.pointInPolygonDirection(l2[0], l2[1], l1[0])
	dir4 := e.pointInPolygonDirection(l2[0], l2[1], l1[1])

	// When intersecting
	if dir1 != dir2 && dir3 != dir4 {
		return true
	}

	// When p2 of line2 are on the line1
	if dir1 == 0 && e.pointInPolygonDirectionOnLine(l1, l2[0]) {
		return true
	}

	// When p1 of line2 are on the line1
	if dir2 == 0 && e.pointInPolygonDirectionOnLine(l1, l2[1]) {
		return true
	}

	//When p2 of line1 are on the line2
	if dir3 == 0 && e.pointInPolygonDirectionOnLine(l2, l1[0]) {
		return true
	}

	// When p1 of line1 are on the line2
	if dir4 == 0 && e.pointInPolygonDirectionOnLine(l2, l1[1]) {
		return true
	}

	return false
}

func (e *NewPolygon) PointInPolygon(p [2]float64) (inside bool) {
	var n = len(e.Loc)

	exLine := [2][2]float64{p, {9999.9, p[Latitude]}}
	count := 0
	i := 0
	for {
		// Forming a line from two consecutive points of poly
		side := [2][2]float64{e.Loc[i], e.Loc[(i+1)%n]}
		if e.pointInPolygonIsIntersect(side, exLine) {
			// If side is intersects exLine
			if e.pointInPolygonDirection(side[0], p, side[1]) == 0 {
				return e.pointInPolygonDirectionOnLine(side, p)
			}

			count++
		}

		i = (i + 1) % n
		if i == 0 {
			break
		}
	}

	if count%2 == 0 {
		return false
	}

	// When count is odd
	return true
}
