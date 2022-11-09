package goosm

import (
	"encoding/binary"
	"time"
)

type Members struct {
	Type string `bson:"type"`
	Ref  int64  `bson:"ref"`
	Role string `bson:"role"`
}

type Relation struct {
	// id do open street maps
	Id             int64   `bson:"id"`
	IdWay          []int64 `bson:"idWay"`
	IdNode         []int64 `bson:"idNode"`
	IdMainRelation []int64 `bson:"idMainRelation"`
	IdRelation     []int64 `bson:"idRelation"`
	IdPolygon      []int64 `bson:"idPolygon"`
	IdPolygonsList []int64 `bson:"idPolygonsList"`
	Version        int64   `bson:"Version"`
	// TimeStamp dentro do Create Street Maps
	TimeStamp time.Time `bson:"timeStamp"`
	// ChangeSet dentro do Create Street Maps
	ChangeSet int64 `bson:"changeSet"`

	Visible bool

	// User Id dentro do Create Street Maps
	UId int64 `bson:"userId"`
	// User Name dentro do Create Street Maps
	User string `bson:"-"`
	// Tags do Create Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag            map[string]string `bson:"tag"`
	International  map[string]string `bson:"international"`
	GeoJSon        string            `bson:"geoJSon"`
	GeoJSonFeature string            `bson:"geoJSonFeature"`

	Members []Members `bson:"-"`
}

func (el *Relation) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}
