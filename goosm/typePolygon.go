package goosm

import (
	"errors"
	"goosm/module/util"
	"log"
	"math"
	"strconv"
)

// English: Polygons are mainly used to demarcate boundaries, and to allow searches in limited areas.
//
// In OpenStreetMaps there is not much difference between a polygon and a way, being a polygon, a way where the first and last point are repeated.
//
// In our case, a polygon can be assembled from a single way, or from the concatenation of several ways distinct.
//
// Português: Polígonos são usados principalmente para demarcar fronteiras e permitir buscas em áreas limitadas.
//
// No OpenStreetMaps não há muita diferença entre um polígono e um way, sendo um polígono um way onde o primeiro e último ponto são repetidos.
//
// No nosso caso, um polígono pode ser montado a partir de um único way ou a partir da concatenação de vários ways distintos.

type Polygon struct {

	// English: id open street maps
	//
	// Português: id do open street maps
	Id int64 `bson:"id"`

	//Surrounding float64 `bson:"surrounding"`

	Visible bool `bson:"bool"`

	// English: Tags OpenStreetMaps
	//
	// The Tags contain all kinds of information, as long as they were imported, the name of a commercial establishment, for example.
	//
	// English: Tags do Create Street Maps
	//
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial, por exemplo.
	Tag map[string]string `bson:"tag"`

	TagFromWay map[string]map[string]string `bson:"tagFromWay"`

	// English: due to some driver mgo limitations, gives less problem if the tags very large, they are separate. By this, the entire content of
	//
	// the tag starts with 'name:.*' it is played within the 'international' key
	//
	// Português: devido a algumas limitações do driver mgo, dá menos problema se as tags muito grandes forem separadas. Por isto, todo conteúdo de
	//
	// tag iniciado com 'name:.*' é jogado dentro da chave 'International'
	International map[string]string `bson:"international"`

	// English: user data
	//
	// Português: dados do usuário
	Data map[string]string `bson:"data"`

	// English: List of polygon forming points
	//
	// Português: Lista dos pontos formadores do polígono
	PointsList []Node `bson:"pointList"`

	// English: The amount of forming points of the polygon
	//
	// Português: Quantidade de pontos formadores do polígono
	Length int `bson:"length"`

	// English: The area of the polygon used for the calculation of the centroid. I do not recommend the use for calculation of the geographic area in this version.
	//
	// Português: Área do polígono usada para o calculo da centroide. Não recomendo o uso para calculo de área geográfica nessa versão.
	Area float64 `bson:"area"`

	// English: Centroid of polygon
	//
	// Português: Centroide do polígono
	Centroid Node `bson:"centroid"`

	// English: used to test if the polygon has been initialized
	//
	// Português: usado para testar se polígono foi inicializado
	Initialize bool `bson:"inicialize"`

	// English: used in the calculations of the point inside the polygon.
	//
	// should not be changed manually
	//
	// Português: usado nos cálculos do ponto dentro do polígono.
	//
	// não deve ser alterado manualmente
	Constant []float64 `bson:"constant"`

	// English: used in the calculations of the point inside the polygon.
	//
	// should not be changed manually
	//
	// Português: usado nos cálculos do ponto dentro do polígono.
	//
	// não deve ser alterado manualmente
	Multiple []float64 `bson:"multiple"`

	// English: distance to the nearest point of the perimeter in the order that points were added
	//
	// Português: distância para o próximo ponto do perímetro na ordem que os pontos foram adicionados
	Distance []Distance `bson:"distance"`

	// English: the total length of the perimeter
	//
	// Português: comprimento total do perímetro
	DistanceTotal Distance `bson:"distanceTotal"`

	// English: ângulo em relação ao próximo ponto do perímetro na ordem que os pontos foram adicionados
	//
	// Português: angle in relation to the next point of the perimeter in the order that points were added
	Angle []Angle `bson:"angle"`

	// English: boundary box in degrees
	//
	// Português: caixa de perímetro em graus decimais
	BBox Box `bson:"bbox"`

	/*
	   bson.M{ "bBoxSearch.1.0": bson.M{ "$gte": -62.162704467 }, "bBoxSearch.1.1": bson.M{ "$lte": -12.341343394 }, "bBoxSearch.0.0": bson.M{ "$lte": -62.162704467 }, "bBoxSearch.0.1": bson.M{ "$gte": -12.341343394 } }
	*/
	//BBoxSearch [2][2]float64 `bson:"bBoxSearch"`

	// English: boundary box in BSon to MongoDB
	//
	// Português: caixa de perímetro em BSon para o MongoDB
	//BBoxBSon       bson.M   `bson:"bBoxBSon"`

	GeoJSonFeature string `bson:"geoJSonFeature"`
	tmp            []Way

	IdWay       []int64         `bson:"idWay"`
	idWayUnique map[int64]int64 `bson:"-"`

	//MinimalDistance Distance `bson:"MinimalDistance" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

//func (el *Polygon) GetIdAsByte() []byte {
//	var ret = make([]byte, 8)
//	binary.LittleEndian.PutUint64(ret, uint64(el.Id))
//
//	return ret
//}

//func (el *Polygon) AsArray() []Polygon {
//	var returnL = make([]Polygon, 1)
//	returnL[0] = *el
//
//	return returnL
//}

//func (el *Polygon) SetMinimalDistance(distanceA Distance) {
//	el.MinimalDistance = distanceA
//}

//func (el *Polygon) SetSurrounding(distance float64) {
//	el.Surrounding = distance
//}

// English: Copies the data of the relation in the polygon.
//
// Português: Copia os dados de uma relação no polígono.
//
// @see dataOfOsm in blog.osm.io

func (el *Polygon) AddRelationDataAsPolygonData(relation *Relation) {
	el.Visible = (*relation).Visible

	el.Tag = (*relation).Tag
	el.International = (*relation).International
}

// English: Copies the data of the way in the polygon.
//
// Português: Copia os dados de uma way no polígono.
//
// @see dataOfOsm in blog.osm.io

func (el *Polygon) AddWayDataAsPolygonData(way *Way) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	//el.Id             =  (*way).Id

	//mapTagLock.Lock()
	//el.Tag            =  (*way).Tag
	//mapTagLock.Unlock()

	//el.International = (*way).International
	//el.Data = (*way).Data
}

// English: Turn a way in a polygon.
//
// Português: Transforma um way em um polígono.

func (el *Polygon) AddWayAsPolygon(way *Way) (err error) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	if len((*way).Tag) > 0 {
		el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag
	}

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	//el.Id             =  (*way).Id

	//mapTagLock.Lock()
	//el.Tag            =  (*way).Tag
	//mapTagLock.Unlock()

	//el.International = (*way).International
	//el.Data = (*way).Data

	for _, loc := range way.Loc {
		el.AddLngLatDegrees(loc[0], loc[1])
	}
	err = el.Init()
	return
}

// English: Adds all ways that are part of a polygon, then process and generate a single polygon
//
// Português: Adiciona todos os ways que fazem parte de um polígono para depois processar e gerar um polígono único

func (el *Polygon) AddWayAsPreProcessingPolygon(way *Way) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	var length = len(way.Loc)
	if length > 0 {
		if way.Loc[0][0] == way.Loc[length-1][0] && way.Loc[0][1] == way.Loc[length-1][1] {
			el.AddWayAsPolygon(way)
			return
		}
	}

	if len(el.tmp) == 0 {
		el.tmp = make([]Way, 0)
	}

	el.tmp = append(el.tmp, *way)
}

// English: Transforms the type PointList in a polygon
//
// Português: Transforma o tipo PointList em um polígono
//func (el *Polygon) SetPointList(pointList PointList) {
//	el.PointsList = pointList.List
//	el.Initialize = false
//}

// English: Transforms the type PointList in a polygon and initializes it
//
// Português: Transforma o tipo PointList em um polígono e inicializa ele
//func (el *Polygon) SetPointListAndInit(pointList PointList) (err error) {
//	el.PointsList = pointList.List
//	el.Initialize = false
//	err = el.Init()
//	return
//}

// English: Adds a point in the format latitude and longitude into decimal degrees to the end of the list of points of the polygon
//
// Português: Adiciona um ponto no formato latitude e longitude em graus decimais ao fim da lista de pontos do polígono

func (el *Polygon) AddLatLngDegrees(latitudeAFlt, longitudeAFlt float64) (err error) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]Node, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	var pointL Node
	err = pointL.SetLngLatDegrees(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return
	}

	//if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
	//	var distanceL = pointL.DistanceBetweenTwoPoints(el.PointsList[len(el.PointsList)-1])
	//	if distanceL >= el.MinimalDistance.Meters {
	//		el.PointsList = append(el.PointsList, pointL)
	//	}
	//	return
	//}

	el.PointsList = append(el.PointsList, pointL)
	return
}

func (el *Polygon) AddPoint(pointA *Node) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]Node, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	//if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
	//	var distanceL = pointA.DistanceBetweenTwoPoints(el.PointsList[len(el.PointsList)-1])
	//	if distanceL >= el.MinimalDistance.Meters {
	//		el.PointsList = append(el.PointsList, *pointA)
	//	}
	//	return
	//}

	el.PointsList = append(el.PointsList, *pointA)
}

// English: Adds a point in the format latitude and longitude into decimal degrees to the top of the list of points of the polygon
//
// Português: Adiciona um ponto no formato latitude e longitude em graus decimais ao início da lista de pontos do polígono

func (el *Polygon) AddLatLngDegreesAtStart(latitudeAFlt, longitudeAFlt float64) (err error) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]Node, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	var pointL Node
	err = pointL.SetLngLatDegrees(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return
	}

	//if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
	//	var distanceL = pointL.DistanceBetweenTwoPoints(el.PointsList[0])
	//	if distanceL >= el.MinimalDistance.Meters {
	//		el.PointsList = append([]Node{pointL}, el.PointsList...)
	//	}
	//	return
	//}

	el.PointsList = append([]Node{pointL}, el.PointsList...)
	return
}

// English: Adds a point in the format longitude and latitude into decimal degrees to the end of the list of points of the polygon
//
// Português: Adiciona um ponto no formato longitude e latitude em graus decimais ao fim da lista de pontos do polígono

func (el *Polygon) AddLngLatDegrees(longitudeAFlt, latitudeAFlt float64) {
	el.AddLatLngDegrees(latitudeAFlt, longitudeAFlt)
}

// English: Adds a point in the format longitude and latitude into decimal degrees to the top of the list of points of the polygon
//
// Português: Adiciona um ponto no formato longitude e latitude em graus decimais ao início da lista de pontos do polígono

func (el *Polygon) AddLngLatDegreesAtStart(longitudeAFlt, latitudeAFlt float64) {
	el.AddLatLngDegreesAtStart(latitudeAFlt, longitudeAFlt)
}

// English: Adds a new key on the tag.
//
// Português: Adiciona uma nova chave na tag do polígono.

func (el *Polygon) AddTag(keyAStr, valueAStr string) {
	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	el.Tag[keyAStr] = valueAStr
}

// English: Initializes the polygon so that the same can be processed at run-time.
//
// Note that this function must be called on each change in the points of the polygon.
//
// Português: Inicializa o polígono para que o mesmo possa ser processado em tempo de execução.
//
// Note que esta função deve ser chamada a cada alteração nos pontos do polígono.

func (el *Polygon) Init() (err error) {
	var distanceList []float64
	var distance float64
	var distanceKey int
	//var distanceAStartBStart float64 = math.MaxFloat64
	//var distanceAStartBEnd float64   = math.MaxFloat64
	//var distanceAEndBStart float64   = math.MaxFloat64
	//var distanceAEndBEnd float64     = math.MaxFloat64
	var k1, k2, lengthStart, lengthTmp int
	//var inverter bool = false
	var pass bool = false
	//var addToEndOfTheSet bool = false
	var pythagorasAStartBStart, pythagorasAStartBEnd, pythagorasAEndBStart, pythagorasAEndBEnd []float64
	var xStart, xEnd, xTmpStart, xTmpEnd float64
	var yStart, yEnd, yTmpStart, yTmpEnd float64

	if len(el.PointsList) < 3 {
		err = errors.New("minimal number of points is 3")
		return
	}

	if el.PointsList[0].Loc[0] != el.PointsList[len(el.PointsList)-1].Loc[0] || el.PointsList[0].Loc[1] != el.PointsList[len(el.PointsList)-1].Loc[1] {
		el.PointsList = append(el.PointsList, el.PointsList[0])
	}

	if len(el.tmp) > 0 {

		for _, loc := range el.tmp[0].Loc {
			el.AddLngLatDegrees(loc[0], loc[1])
		}

		pythagorasAStartBStart = make([]float64, len(el.tmp))
		pythagorasAStartBEnd = make([]float64, len(el.tmp))
		pythagorasAEndBStart = make([]float64, len(el.tmp))
		pythagorasAEndBEnd = make([]float64, len(el.tmp))

		distanceList = make([]float64, len(el.tmp))

		for k2 = 1; k2 != len(el.tmp); k2 += 1 {
			distance = math.MaxFloat64

			for k1 = 0; k1 != len(el.tmp); k1 += 1 {
				distanceList[k1] = math.MaxFloat64
			}

			lengthStart = len(el.PointsList) - 1

			xStart = el.PointsList[0].Loc[0]
			yStart = el.PointsList[0].Loc[1]
			xEnd = el.PointsList[lengthStart].Loc[0]
			yEnd = el.PointsList[lengthStart].Loc[1]

			for k1 = 1; k1 != len(el.tmp); k1 += 1 {

				lengthTmp = len(el.tmp[k1].Loc) - 1

				xTmpStart = el.tmp[k1].Loc[0][0]
				yTmpStart = el.tmp[k1].Loc[0][1]
				xTmpEnd = el.tmp[k1].Loc[lengthTmp][0]
				yTmpEnd = el.tmp[k1].Loc[lengthTmp][1]

				pythagorasAStartBStart[k1] = util.Pythagoras(xStart, yStart, xTmpStart, yTmpStart)
				pythagorasAStartBEnd[k1] = util.Pythagoras(xStart, yStart, xTmpEnd, yTmpEnd)
				pythagorasAEndBStart[k1] = util.Pythagoras(xEnd, yEnd, xTmpStart, yTmpStart)
				pythagorasAEndBEnd[k1] = util.Pythagoras(xEnd, yEnd, xTmpEnd, yTmpEnd)

				distanceList[k1] = math.Min(distanceList[k1], pythagorasAStartBStart[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAStartBEnd[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAEndBStart[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAEndBEnd[k1])

				if distanceList[k1] < distance {
					distance = distanceList[k1]
					distanceKey = k1
				}
			}

			if pythagorasAStartBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAStartBStart[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAStartBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] {
				pass = true
				for _, loc := range el.tmp[distanceKey].Loc {
					el.AddLngLatDegreesAtStart(loc[0], loc[1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAEndBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAEndBEnd[distanceKey] < pythagorasAStartBEnd[distanceKey] {
				pass = true
				for _, loc := range el.tmp[distanceKey].Loc {
					el.AddLngLatDegrees(loc[0], loc[1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAStartBEnd[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAEndBEnd[distanceKey] < pythagorasAEndBStart[distanceKey] {
				pass = true
				//for _, loc := range el.tmp[k1].Loc {
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegreesAtStart(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] && pythagorasAStartBEnd[distanceKey] < pythagorasAEndBEnd[distanceKey] && pythagorasAEndBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] {
				pass = true
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegrees(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			} else {
				pass = true
				//for _, loc := range el.tmp[k1].Loc {
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegrees(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			}

			if pass == true {
				log.Fatalf("")
			}
		}
	}

	if len(el.PointsList) == 0 {
		return errors.New("polygon has't points")
	}

	el.Initialize = true

	if el.PointsList[0].Loc[0] != el.PointsList[len(el.PointsList)-1].Loc[0] || el.PointsList[0].Loc[1] != el.PointsList[len(el.PointsList)-1].Loc[1] {
		el.PointsList = append(el.PointsList, el.PointsList[0])
	}

	el.Constant = make([]float64, len(el.PointsList))
	el.Multiple = make([]float64, len(el.PointsList))
	el.Length = len(el.PointsList)

	lastCornerLUInt := len(el.PointsList) - 1

	for i := 0; i != len(el.PointsList); i += 1 {

		if el.PointsList[lastCornerLUInt].getLatitudeAsRadians() == el.PointsList[i].getLatitudeAsRadians() {
			el.Constant[i] = el.PointsList[i].getLongitudeAsRadians()
			el.Multiple[i] = 0
		} else {
			el.Constant[i] = el.PointsList[i].getLongitudeAsRadians() -
				(el.PointsList[i].getLatitudeAsRadians()*el.PointsList[lastCornerLUInt].getLongitudeAsRadians())/
					(el.PointsList[lastCornerLUInt].getLatitudeAsRadians()-el.PointsList[i].getLatitudeAsRadians()) +
				(el.PointsList[i].getLatitudeAsRadians()*el.PointsList[i].getLongitudeAsRadians())/
					(el.PointsList[lastCornerLUInt].getLatitudeAsRadians()-el.PointsList[i].getLatitudeAsRadians())
			el.Multiple[i] = (el.PointsList[lastCornerLUInt].getLongitudeAsRadians() -
				el.PointsList[i].getLongitudeAsRadians()) /
				(el.PointsList[lastCornerLUInt].getLatitudeAsRadians() - el.PointsList[i].getLatitudeAsRadians())
		}

		lastCornerLUInt = i
	}

	el.centroid()
	el.area()

	var pointA = Node{}
	var pointB = Node{}
	var distanceListLA []Distance
	var distanceL = Distance{}
	distanceL.SetMeters(0.0)

	var k int

	var angleList []Angle
	var angle = Angle{}
	angle.SetDegrees(0.0)

	distanceListLA = make([]Distance, len(el.PointsList))
	distanceListLA[0] = distanceL

	angleList = make([]Angle, len(el.PointsList))

	for keyRefLInt64 := range el.PointsList {
		if keyRefLInt64 != 0 {
			err = pointA.SetLngLatRadians(el.PointsList[keyRefLInt64-1].Rad[0], el.PointsList[keyRefLInt64-1].Rad[1])
			if err != nil {
				return
			}

			err = pointB.SetLngLatRadians(el.PointsList[keyRefLInt64].Rad[0], el.PointsList[keyRefLInt64].Rad[1])
			if err != nil {
				return
			}

			angleList[keyRefLInt64-1].SetDegrees(pointA.DirectionBetweenTwoPoints(pointB))

			distanceListLA[keyRefLInt64].SetMeters(pointA.DistanceBetweenTwoPoints(pointB))
			distanceL.AddMeters(distanceListLA[keyRefLInt64].GetMeters())

			k = keyRefLInt64
		}
		angleList[k].SetDegrees(pointA.DirectionBetweenTwoPoints(pointB))
	}

	el.Distance = distanceListLA
	el.DistanceTotal = distanceL
	el.Angle = angleList
	el.BBox, err = GetBox(&el.PointsList)
	//el.BBoxBSon = GetBSonBoxInDegrees(&el.PointsList)
	//el.BBoxSearch = [2][2]float64{el.BBox.UpperRight.Loc, el.BBox.BottomLeft.Loc}

	return
}

// English: Tests if the point is contained within the polygon.
//
// If the point is above the line of the edge, the same can give a response undetermined because of the lease of the decimals.
//
// Português: Testa se o ponto está contido dentro do polígono.
//
// Se o ponto estiver em cima da linha da borda, o mesmo pode dá uma resposta indeterminada devido ao arrendamento das casas decimais

func (el *Polygon) PointInPolygon(pointA Node) (yesInPolygon bool, err error) {
	if el.Initialize == false {
		el.Initialize = true
		err = el.Init()
		if err != nil {
			return
		}
	}

	lastCornerLUInt := len(el.PointsList) - 1
	tempLBoo := false

	for i := 0; i != len(el.PointsList); i += 1 {
		if el.PointsList[i].getLatitudeAsRadians() < pointA.getLatitudeAsRadians() &&
			el.PointsList[lastCornerLUInt].getLatitudeAsRadians() >= pointA.getLatitudeAsRadians() ||
			el.PointsList[lastCornerLUInt].getLatitudeAsRadians() < pointA.getLatitudeAsRadians() &&
				el.PointsList[i].getLatitudeAsRadians() >= pointA.getLatitudeAsRadians() {

			tempLBoo = pointA.getLatitudeAsRadians()*el.Multiple[i]+el.Constant[i] < pointA.getLongitudeAsRadians()

			// oddNodesLBoo = ( oddNodesLBoo XOR tempLBoo )
			yesInPolygon = (yesInPolygon || tempLBoo) && !(yesInPolygon && tempLBoo)
		}
		lastCornerLUInt = i
	}

	return
}

func (el *Polygon) centroid() {
	el.Centroid.Loc = [2]float64{0.0, 0.0}
	el.Centroid.Rad = [2]float64{0.0, 0.0}

	var areaLFlt = 0.0
	var a = 0.0

	var i = 0
	for ; i != len(el.PointsList)-1; i += 1 {
		a = el.PointsList[i].Rad[0]*el.PointsList[i+1].Rad[1] - el.PointsList[i+1].Rad[0]*el.PointsList[i].Rad[1]
		areaLFlt += a
		el.Centroid.Rad[0] += (el.PointsList[i].Rad[0] + el.PointsList[i+1].Rad[0]) * a
		el.Centroid.Rad[1] += (el.PointsList[i].Rad[1] + el.PointsList[i+1].Rad[1]) * a
	}

	a = el.PointsList[i].Rad[0]*el.PointsList[0].Rad[1] - el.PointsList[0].Rad[0]*el.PointsList[i].Rad[1]
	areaLFlt += a
	el.Centroid.Rad[0] += (el.PointsList[i].Rad[0] + el.PointsList[0].Rad[0]) * a
	el.Centroid.Rad[1] += (el.PointsList[i].Rad[1] + el.PointsList[0].Rad[1]) * a

	areaLFlt *= 0.5
	el.Centroid.Rad[0] /= 6.0 * areaLFlt
	el.Centroid.Rad[1] /= 6.0 * areaLFlt

	el.Centroid.Loc[0] = util.RadiansToDegrees(el.Centroid.Rad[0])
	el.Centroid.Loc[1] = util.RadiansToDegrees(el.Centroid.Rad[1])
}

func (el *Polygon) area() {
	el.Area = 0.0

	var polygonL Polygon

	polygonL.PointsList = make([]Node, len(el.PointsList))

	for i := 0; i != len(el.PointsList); i += 1 {
		earthRadiusL := EarthRadius(el.PointsList[i])
		earthRadiusLFlt := earthRadiusL.GetMeters() * 1000
		_ = polygonL.PointsList[i].SetLngLatRadians(el.PointsList[i].Rad[1]*earthRadiusLFlt, el.PointsList[i].Rad[0]*earthRadiusLFlt)
	}

	var i = 0
	for ; i != len(polygonL.PointsList)-1; i += 1 {
		el.Area += polygonL.PointsList[i].Rad[0]*polygonL.PointsList[i+1].Rad[1] - polygonL.PointsList[i+1].Rad[0]*polygonL.PointsList[i].Rad[1]
	}

	el.Area += polygonL.PointsList[i].Rad[0]*polygonL.PointsList[0].Rad[1] - polygonL.PointsList[0].Rad[0]*polygonL.PointsList[i].Rad[1]
	el.Area *= 0.5
}

// English: Determines the box in which the polygon is contained to be used with the function $box of Mongo DB.
//
// For higher performance of the database, use this function to grab all the points within the box, and then use the function PointInPolygon() to test.
//
// If the point is not in the box, it is not within the polygon.
//
// The answer will be in radians.
//
// Português: Determina a caixa onde o polígono está contido para ser usado com a função $box do MongoDB.
//
// Para maior desempenho do banco, use esta função para pegar todos os pontos dentro da caixa e depois use a função PointInPolygon() para testar.
//
// Caso o ponto não esteja na caixa, ele não está dentro do polígono.
//
// A resposta será em radianos.

func (el *Polygon) GetBox() Box { return el.BBox }

// English: Determines the box in which the polygon is contained to be used with the function $box of Mongo DB.
//
// For higher performance of the database, use this function to grab all the points within the box, and then use the function PointInPolygon() to test.
//
// If the point is not in the box, it is not within the polygon.
//
// The answer will be in decimal degrees.
//
// Português: Determina a caixa onde o polígono está contido para ser usado com a função $box do MongoDB.
//
// Para maior desempenho do banco, use esta função para pegar todos os pontos dentro da caixa e depois use a função PointInPolygon() para testar.
//
// Caso o ponto não esteja na caixa, ele não está dentro do polígono.
//
// A resposta será em graus decimais.
//func (el *Polygon) GetBSonBoxInDegrees() bson.M { return el.BBoxBSon }

// English: Mounts the geoJSon Feature from polygon and populates the key GeoJSonFeature into the struct
//
// Português: Monta o geoJSon Feature do polígono e popula a chave GeoJSonFeature na struct

func (el *Polygon) MakeGeoJSonFeature() string {

	// fixme: fazer
	//if el.Id == 0 {
	//	el.Id = util.AutoId.Get(el.DbCollectionName)
	//}

	var geoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygon(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

// English: Resize a polygon based on the distance between the centroide and the points of construction of the same.
//
// There may be distortions in relation to the polygon and the original polygon.
//
// This function should be improved in the future.
//
// Português: Redimensiona um poligono baseado na distância entre a centroide e os pontos de construção do mesmo.
//
// Pode haver distorções em relação ao polígono original.
//
// Esta função deve ser melhorada em um futuro.

func (el *Polygon) Resize(distanceAObj Distance) (polygon Polygon, err error) {
	if el.Initialize == false {
		el.Initialize = true
		err = el.Init()
		if err != nil {
			return
		}
	}

	polygon.PointsList = make([]Node, len(el.PointsList))

	distance := Distance{}

	direction := Angle{}
	point := Node{}

	for k, v := range el.PointsList {
		distance.SetMeters(v.DistanceBetweenTwoPoints(el.Centroid))
		distance.AddMeters(distanceAObj.GetMeters())
		direction.SetDegrees(v.DirectionBetweenTwoPoints(el.Centroid))
		direction.AddDegrees(180)
		point, err = el.Centroid.DestinationPoint(distance.GetMeters(), direction.GetAsDegrees())
		if err != nil {
			return
		}

		polygon.PointsList[k] = point
	}

	err = polygon.Init()
	if err != nil {
		return
	}

	return
}

// English: Returns the length of the line of the perimeter
//
// Português: Devolve o comprimento da linha de perímetro

func (el *Polygon) GetRadius() (distance Distance, err error) {
	if el.Initialize == false {
		el.Initialize = true
		err = el.Init()
		if err != nil {
			return
		}
	}

	var distanceTmp Distance

	distance.SetMeters(0.0)

	for _, v := range el.PointsList {
		distanceTmp.SetMeters(v.DistanceBetweenTwoPoints(el.Centroid))
		distance.SetMetersIfGreaterThan(distanceTmp.GetMeters())
	}

	return
}

// English: Converts the polygon in a Convex Hull
//
// Special thanks to Valeriy Streltsov for his work in C++
//
// Português: Converte o polígono em um Convex Hull
//
// Agradecimento especial ao Valeriy Streltsov pelo seu trabalho em C++
//func (el *Polygon) ConvertToConvexHull() {
//	var points = PointList{}
//	points.List = el.PointsList
//	points = points.ConvexHull()
//
//	el.PointsList = points.List
//}

// English: Converts the polygon in a Concave Hull
//
// Special thanks to Valeriy Streltsov for his work in C++
//
// Português: Converte o polígono em um Concave Hull
//
// Agradecimento especial ao Valeriy Streltsov pelo seu trabalho em C++
//func (el *Polygon) ConvertToConcaveHull(n float64) (err error) {
//	var points = PointList{}
//	points.List = el.PointsList
//	points, err = points.ConcaveHull(n)
//	if err != nil {
//		return
//	}
//
//	el.PointsList = points.List
//	el.GeoJSonFeature = ""
//	el.Init()
//	el.MakeGeoJSonFeature()
//	return
//}

/*func (el *Polygon) FindPointInPolygon(pointQueryAObj bson.M) (error, PointList) {
  return el.FindPointInOnePolygon(bson.M{}, pointQueryAObj)
}

func (el *Polygon) FindPointInOnePolygon(polygonQueryAObj bson.M, pointQueryAObj bson.M) (error, PointList) {
  var err error
  var pointListL PointList = PointList{}

  if el.DbCollectionName == "" {
    el.Prepare()
  }

  pointListL = PointList{
    DbCollectionName: el.DbCollectionNameForNode,
  }

  var returnPointsL PointList = PointList{}
  returnPointsL.List = make([]Node, 0)

  if !reflect.DeepEqual(polygonQueryAObj, bson.M{}) {
    err = el.MongoFindOne(polygonQueryAObj)
    if err != nil {
      return err, returnPointsL
    }
  }

  if el.Initialize == false {
    el.Init()
  }

  if el.Id == 0 {
    return nil, returnPointsL
  }

  if !reflect.DeepEqual(pointQueryAObj, bson.M{}) {
    pointQueryAObj = bson.M{`$and`: []bson.M{{`loc`: el.BBoxBSon}, pointQueryAObj}}
  } else {
    pointQueryAObj = bson.M{`loc`: el.BBoxBSon}
  }

  err = pointListL.MongoFind(pointQueryAObj)
  if err != nil {
    log.Criticalf("gOsm.geoMath.geoTypePolygon.Error: ", err.Error())
    return err, returnPointsL
  }

  for _, pointToTestL := range pointListL.List {
    if el.PointInPolygon(pointToTestL) == true {
      pointToTestL.MongoFindOne(bson.M{"id": pointToTestL.Id})
      returnPointsL.List = append(returnPointsL.List, pointToTestL)
    }
  }

  return nil, returnPointsL
}*/
