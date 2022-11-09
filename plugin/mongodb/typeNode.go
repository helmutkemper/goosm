package mongodb

import "goosm/goosm"

type Node struct {
	Id int64 `bson:"_id"`

	// Array de localização geográfica.
	// [0:x:longitude,1:y:latitude]
	// Este campo deve obrigatoriamente ser um array devido a indexação do MongoDB
	Loc GeoJSonPoint `bson:"loc"`

	// Tags do Create Street Maps
	Tag map[string]string `bson:"tag,omitempty"`

	GeoJSonFeature string `bson:"geoJSonFeature,omitempty"`
}

func (e Node) ToOsmNode() (node goosm.Node) {
	node.Id = e.Id
	node.Tag = e.Tag
	node.Loc = e.Loc.Coordinates
	node.GeoJSonFeature = e.GeoJSonFeature
	return
}

func (e *Node) ToDbNode(node *goosm.Node) (dbNode Node) {
	e.Id = node.Id
	e.Tag = node.Tag
	e.Loc.Type = "Point"
	e.Loc.Coordinates = node.Loc
	e.GeoJSonFeature = node.GeoJSonFeature
	return
}
