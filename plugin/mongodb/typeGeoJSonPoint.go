package mongodb //nolint:typecheck

type GeoJSonPoint struct {
	Type        string     `bson:"type"`
	Coordinates [2]float64 `bson:"coordinates"`
}
