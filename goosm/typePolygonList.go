package goosm

import (
	"encoding/binary"
	"strconv"
	"time"
)

type PolygonList struct {
	// id do open street maps
	Id int64 `bson:"id"`
	// Versão dentro do Create Street Maps
	Version int64 `bson:"version"`
	// TimeStamp dentro do Create Street Maps
	TimeStamp time.Time `bson:"timeStamp"`
	// ChangeSet dentro do Create Street Maps
	ChangeSet int64 `bson:"changeSet"`

	Visible bool `bson:"visible"`

	// User Id dentro do Create Street Maps
	UId int64 `bson:"userId"`
	// User Name dentro do Create Street Maps
	User string `bson:"-"`
	// Tags do Create Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag           map[string]string            `bson:"tag"`
	TagFromWay    map[string]map[string]string `bson:"tagFromWay"`
	International map[string]string            `bson:"international"`
	Initialized   bool                         `bson:"inicialize"`
	// Dados do usuário
	// Como o GO é fortemente tipado, eu obtive problemas em estender o struct de forma satisfatória e permitir ao usuário
	// do sistema gravar seus próprios dados, por isto, este campo foi criado. Use-o a vontade.
	Data map[string]string `bson:"data"`
	Role string            `bson:"role"`

	idRelationUnique map[int64]int64 `bson:"-"`
	idPolygonUnique  map[int64]int64 `bson:"-"`
	idWayUnique      map[int64]int64 `bson:"-"`

	IdRelation []int64   `bson:"idRelation"`
	IdPolygon  []int64   `bson:"idPolygon"`
	IdWay      []int64   `bson:"idWay"`
	List       []Polygon `bson:"list"`
	// en: boundary box in degrees
	// pt: caixa de perímetro em graus decimais
	BBox           Box    `bson:"bbox"`
	GeoJSon        string `bson:"geoJSon"`
	GeoJSonFeature string `bson:"geoJSonFeature"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *PolygonList) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

// en: Copies the data of the relation in the polygon.
//
// pt: Copia os dados de uma relação no polígono.
//
// @see dataOfOsm in blog.osm.io
func (el *PolygonList) AddRelationDataAsPolygonData(relation *Relation) {
	if len(el.IdRelation) == 0 {
		el.IdRelation = make([]int64, 0)
	}
	if len(el.idRelationUnique) == 0 {
		el.idRelationUnique = make(map[int64]int64)
	}
	if el.idRelationUnique[(*relation).Id] != (*relation).Id {
		el.idRelationUnique[(*relation).Id] = (*relation).Id
		el.IdRelation = append(el.IdRelation, (*relation).Id)
	}

	el.Version = (*relation).Version
	el.TimeStamp = (*relation).TimeStamp
	el.ChangeSet = (*relation).ChangeSet
	el.Visible = (*relation).Visible
	el.UId = (*relation).UId
	el.User = (*relation).User
	el.Tag = (*relation).Tag
	el.International = (*relation).International
}

func (el *PolygonList) AddWayAsPolygon(way *Way) {
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

	var polygon = Polygon{}
	polygon.AddWayAsPolygon(way)
	polygon.Init()

	el.AddPolygon(&polygon)
}

func (el *PolygonList) AddPolygon(polygon *Polygon) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}
	el.TagFromWay[strconv.FormatInt(polygon.Id, 10)] = polygon.Tag

	if len(el.IdPolygon) == 0 {
		el.IdPolygon = make([]int64, 0)
	}
	if len(el.idPolygonUnique) == 0 {
		el.idPolygonUnique = make(map[int64]int64)
	}
	if el.idPolygonUnique[(*polygon).Id] != (*polygon).Id {
		el.idPolygonUnique[(*polygon).Id] = (*polygon).Id
		el.IdPolygon = append(el.IdPolygon, polygon.Id)
	}

	if len(el.List) == 0 {
		el.List = make([]Polygon, 0)
	}

	el.List = append(el.List, *polygon)
}

/*func ( el *PolygonList ) FindFromPolygon( queryAObj bson.M ) error {
  err := el.DbStt.TestConnection()
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
    return err
  }

  err = el.DbStt.Find( el.DbCollectionName, &el.List, queryAObj )
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
    return err
  }

  var polygonTmp = PolygonStt{}

  for k, polygon := range el.List {
    polygonTmp = PolygonStt{
      DbCollectionName: el.DbCollectionName,
    }

    err = polygonTmp.FromGrid( polygon.Id )
    if err != nil{
      log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
      return err
    }

    el.List[ k ] = polygonTmp
  }

  return err
}*/

/*func ( el *PolygonList ) FindPolygonByLngLatDegrees( lng, lat float64 ) error {
  var err error
  var polygonList = PolygonList{}
  var point = PointStt{}
  point.SetLngLatDegrees( lng, lat )

  if el.DbCollectionName == "" {
    el.Prepare()
  }

  err = el.DbStt.Find( el.DbCollectionNameForPolygon, &el.List, bson.M{ "bBoxSearch.1.0": bson.M{ "$gte": lng }, "bBoxSearch.1.1": bson.M{ "$lte": lat }, "bBoxSearch.0.0": bson.M{ "$lte": lng }, "bBoxSearch.0.1": bson.M{ "$gte": lat } } )
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err.Error() )
    return err
  }

  for _, polygon := range el.List{
    if polygon.HasKeyValue == true {
      err = polygon.DbKeyValueFind(polygon.Id)
      if err != nil {
        log.Criticalf("gOsm.geoMath.el.error: %s", err.Error())
        return err
      }
    }

    if polygon.PointInPolygon( point ) {
      polygonList.AddPolygon( &polygon )
    }
  }

  el.List = make( []PolygonStt, 0 )
  for _, polygon := range polygonList.List{
    el.List = append( el.List, polygon )
  }

  return nil
}*/

func (el *PolygonList) MakeGeoJSonAllFeatures() string {

	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS )
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringAllFeatures()

	return el.GeoJSonFeature
}

func (el *PolygonList) MakeGeoJSon() string {
	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS )
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSon, _ = geoJSon.String()

	return el.GeoJSon
}

func (el *PolygonList) MakeGeoJSonFeature() string {
	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS ) //fixme: o que esta constante está fazendo aqui nesse arquivo?
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

func (el *PolygonList) Initialize() (err error) {
	el.Initialized = true

	el.BBox, err = GetBoxPolygonList(el)
	return
}

//func (el *PolygonList) ConvertToConvexHull() {
//	var pList PointList = PointList{}
//	pList.List = make([]Node, 0)
//
//	for _, polygon := range el.List {
//		for _, point := range polygon.PointsList {
//			pList.List = append(pList.List, point)
//		}
//	}
//
//	var hull PointList = pList.ConvexHull()
//	el.List = make([]Polygon, 1)
//	el.List[0].PointsList = hull.List
//}

//func (el *PolygonList) ConvertToConcaveHull(n float64, err error) {
//	var pList PointList = PointList{}
//	pList.List = make([]Node, 0)
//
//	for _, polygon := range el.List {
//		for _, point := range polygon.PointsList {
//			pList.List = append(pList.List, point)
//		}
//	}
//
//	var hull PointList
//	hull, err = pList.ConcaveHull(n)
//	if err != nil {
//		return
//	}
//
//	el.List = make([]Polygon, 1)
//	el.List[0].PointsList = hull.List
//}
