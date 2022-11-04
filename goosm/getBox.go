package goosm

import (
	"math"
)

// GetBox
//
// English:
//
// Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the points
// contained within a rectangular box
//
// Português:
//
// Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos
// dentro de uma caixa retangular
func GetBox(list *[]Node) (box Box, err error) {
	var latMin, latMax, lngMin, lngMax float64

	for k, v := range *list {
		if k == 0 {
			latMin = v.getLatitudeAsRadians()
			latMax = v.getLatitudeAsRadians()

			lngMin = v.getLongitudeAsRadians()
			lngMax = v.getLongitudeAsRadians()
		} else {
			latMin = math.Min(latMin, v.getLatitudeAsRadians())
			latMax = math.Max(latMax, v.getLatitudeAsRadians())

			lngMin = math.Min(lngMin, v.getLongitudeAsRadians())
			lngMax = math.Max(lngMax, v.getLongitudeAsRadians())
		}
	}

	err = box.UpperRight.SetLngLatRadians(lngMin, latMax)
	if err != nil {
		return
	}
	box.UpperRight.MakeGeoJSonFeature()

	err = box.BottomLeft.SetLngLatRadians(lngMax, latMin)
	if err != nil {
		return
	}
	box.BottomLeft.MakeGeoJSonFeature()

	return
}

// GetBoxFlt
//
// English:
//
// Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the
// points contained within a rectangular box.
//
// Português:
//
// Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos
// dentro de uma caixa retangular
func GetBoxFlt(list *[][2]float64) (box Box, err error) {
	var latMin, latMax, lngMin, lngMax float64

	for k, v := range *list {
		if k == 0 {
			latMin = v[1]
			latMax = v[1]

			lngMin = v[0]
			lngMax = v[0]
		} else {
			latMin = math.Min(latMin, v[1])
			latMax = math.Max(latMax, v[1])

			lngMin = math.Min(lngMin, v[0])
			lngMax = math.Max(lngMax, v[0])
		}
	}

	err = box.BottomLeft.SetLngLatDegrees(lngMax, latMin)
	if err != nil {
		return
	}
	box.BottomLeft.MakeGeoJSonFeature()

	err = box.UpperRight.SetLngLatDegrees(lngMin, latMax)
	if err != nil {
		return
	}
	box.UpperRight.MakeGeoJSonFeature()

	return
}

// GetBoxList
//
// English:
//
// Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the
// points contained within a rectangular box.
//
// Português:
//
// Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos
// dentro de uma caixa retangular
//func GetBoxList(list *[]PointList) (box Box, err error) {
//	var latMin, latMax, lngMin, lngMax float64
//
//	for sk, subList := range *list {
//		for k, v := range subList.List {
//			if k == 0 && sk == 0 {
//				latMin = v.getLatitudeAsRadians()
//				latMax = v.getLatitudeAsRadians()
//
//				lngMin = v.getLatitudeAsRadians()
//				lngMax = v.getLatitudeAsRadians()
//			} else {
//				latMin = math.Min(latMin, v.getLatitudeAsRadians())
//				latMax = math.Max(latMax, v.getLatitudeAsRadians())
//
//				lngMin = math.Min(lngMin, v.getLatitudeAsRadians())
//				lngMax = math.Max(lngMax, v.getLatitudeAsRadians())
//			}
//		}
//	}
//
//	err = box.BottomLeft.SetLngLatRadians(lngMin, latMax)
//	if err != nil {
//		return
//	}
//	box.BottomLeft.MakeGeoJSonFeature()
//
//	err = box.UpperRight.SetLngLatRadians(lngMax, latMin)
//	if err != nil {
//		return
//	}
//	box.UpperRight.MakeGeoJSonFeature()
//
//	return
//}

// GetBoxPolygonList
//
// English:
//
// Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the
// points contained within a rectangular box.
//
// Português:
//
// Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos
// dentro de uma caixa retangular
func GetBoxPolygonList(list *PolygonList) (box Box, err error) {
	var latMin, latMax, lngMin, lngMax float64

	for sk, subList := range list.List {
		for k, v := range subList.PointsList {
			if k == 0 && sk == 0 {
				latMin = v.getLatitudeAsRadians()
				latMax = v.getLatitudeAsRadians()

				lngMin = v.getLatitudeAsRadians()
				lngMax = v.getLatitudeAsRadians()
			} else {
				latMin = math.Min(latMin, v.getLatitudeAsRadians())
				latMax = math.Max(latMax, v.getLatitudeAsRadians())

				lngMin = math.Min(lngMin, v.getLatitudeAsRadians())
				lngMax = math.Max(lngMax, v.getLatitudeAsRadians())
			}
		}
	}

	err = box.BottomLeft.SetLngLatRadians(lngMin, latMax)
	if err != nil {
		return
	}
	box.BottomLeft.MakeGeoJSonFeature()

	err = box.UpperRight.SetLngLatRadians(lngMax, latMin)
	if err != nil {
		return
	}
	box.UpperRight.MakeGeoJSonFeature()

	return
}
