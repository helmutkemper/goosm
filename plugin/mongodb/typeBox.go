package mongodb

type Box struct {
	BottomLeft Node `bson:"bottomleft"`
	UpperRight Node `bson:"upperright"`
}
