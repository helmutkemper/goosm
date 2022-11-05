package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"goosm/constants"
	"goosm/goosm"
	"math"
	"strings"
	"time"
)

type Osm struct {
	timeout        time.Duration
	Client         *mongo.Client
	CancelFunc     context.CancelFunc
	ClientNode     *mongo.Collection
	ClientWrongWay *mongo.Collection
	ClientWay      *mongo.Collection
}

func (e *Osm) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

func (e *Osm) Connect(connectionString string, _ ...interface{}) (err error) {
	e.Client, err = mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Connect(ctx)
	cancel()
	if err != nil {
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Ping(ctx, readpref.Primary())
	cancel()
	return
}

func (e *Osm) Close() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Disconnect(ctx)
	cancel()
	return
}

func (e *Osm) New() (referenceInitialized interface{}, err error) {
	if err = e.Connect(constants.KMongoDBConnectionString); err != nil {
		return
	}

	if err = e.createTableNode(); err != nil {
		return
	}

	if err = e.createTableWay(); err != nil {
		return
	}

	if err = e.createTableWrongWay(); err != nil {
		return
	}

	return e, err
}

func (e *Osm) SetNodeOne(node *goosm.Node) (err error) {
	var nodeDb = Node{}
	nodeDb.ToDbNode(node)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.ClientNode.InsertOne(ctx, nodeDb)
	cancel()
	return
}

func (e *Osm) SetNodeMany(list *[]goosm.Node) (err error) {
	nodeDb := Node{}
	var listDb = make([]interface{}, len(*list))
	for key, node := range *list {
		nodeDb.ToDbNode(&node)
		listDb[key] = nodeDb
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.ClientNode.InsertMany(ctx, listDb)
	cancel()
	return
}

func (e *Osm) SetWrongWayNodeMany(list *[]goosm.Node) (err error) {
	nodeDb := Node{}
	var listDb = make([]interface{}, len(*list))
	for key, node := range *list {
		nodeDb.ToDbNode(&node)
		listDb[key] = nodeDb
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.ClientWrongWay.InsertMany(ctx, listDb)
	cancel()
	return
}

func (e *Osm) GetNodeById(id int64) (node goosm.Node, err error) {
	var nodeDb Node
	e.ClientNode = e.Client.Database(constants.KMongoDBDatabase).Collection(constants.KMongoDBCollectionNode)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.ClientNode.FindOne(ctx, bson.M{"_id": id}).Decode(&nodeDb)
	cancel()
	if err != nil {
		return
	}

	node = nodeDb.Node()
	return
}

func (e *Osm) wayJoinLast(loc [2]float64, timeout time.Duration, idListGarbage *[]int64, name string) (way Way, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = e.ClientWay.FindOne(ctx, bson.M{"tag.name": name, "_id": bson.M{"$nin": *idListGarbage}, "$or": []bson.M{{"locFirst": loc}, {"locLast": loc}}}).Decode(&way)
	cancel()
	if err != nil {
		return
	}

	if way.Id == 0 {
		return
	}

	*idListGarbage = append(*idListGarbage, way.Id)
	return
}

func (e *Osm) wayJoinTag(loc [2]float64, key string, value interface{}, timeout time.Duration, idListGarbage *[]int64, name string) (way Way, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	query := bson.M{"$and": []bson.M{{key: value}, {"_id": bson.M{"$nin": *idListGarbage}}, {"$or": []bson.M{{"locFirst": loc}, {"locLast": loc}}}}}
	err = e.ClientWay.FindOne(ctx, query).Decode(&way)
	cancel()
	if err != nil {
		return
	}

	if way.Id == 0 {
		return
	}

	*idListGarbage = append(*idListGarbage, way.Id)
	return
}

// Float64Data is a named type for []float64 with helper methods
type Float64Data []float64

// Len returns length of slice
func (f Float64Data) Len() int { return len(f) }

// Sum returns the total of all the numbers in the data
func (f Float64Data) Sum() (float64, error) { return Sum(f) }

// Sum adds all the numbers of a slice together
func Sum(input Float64Data) (sum float64, err error) {

	if input.Len() == 0 {
		return math.NaN(), errors.New("empty input")
	}

	// Add em up
	for _, n := range input {
		sum += n
	}

	return sum, nil
}

func PopulationVariance(input Float64Data) (pvar float64, err error) {

	v, err := _variance(input, 0)
	if err != nil {
		return math.NaN(), err
	}

	return v, nil
}

// StandardDeviationPopulation finds the amount of variation from the population
func StandardDeviationPopulation(input Float64Data) (sdev float64, err error) {

	if input.Len() == 0 {
		return math.NaN(), errors.New("empty input")
	}

	// Get the population variance
	vp, _ := PopulationVariance(input)

	// Return the population standard deviation
	return math.Sqrt(vp), nil
}

// _variance finds the variance for both population and sample data
func _variance(input Float64Data, sample int) (variance float64, err error) {

	if input.Len() == 0 {
		return math.NaN(), errors.New("empty input")
	}

	// Sum the square of the mean subtracted from each number
	m, _ := Mean(input)

	for _, n := range input {
		variance += (n - m) * (n - m)
	}

	// When getting the mean of the squared differences
	// "sample" will allow us to know if it's a sample
	// or population and wether to subtract by one or not
	return variance / float64(input.Len()-(1*sample)), nil
}

// Mean gets the average of a slice of numbers
func Mean(input Float64Data) (float64, error) {

	if input.Len() == 0 {
		return math.NaN(), errors.New("empty input")
	}

	sum, _ := input.Sum()

	return sum / float64(input.Len()), nil
}

func (e *Osm) angle(a float64) float64 {
	if a >= 180 {
		a -= 180
	}
	return a
}

// WayJoinGeoJSonFeatures
//
// English:
//
// Joins all the forming segments of a way that has the "tag.name" property defined and returns the geoJSon of the same
//
//	Input:
//	  id: ID of any segment forming the set;
//	  timeout: timeout of each database request.
//
//	Output:
//	  distanceMeters: total distance in meters (forks are counted);
//	  features: list of geojson features;
//	  err: golang's default error object.
//
// Português:
//
// Junta todos os seguimentos formadores de um way que tenha a propriedade "tag.name" definida e retorna o geoJSon do
// mesmo
//
//	Entrada:
//	  id: ID de qualquer seguimento formador do conjunto;
//	  timeout: tempo limit de cada requisição do banco de dados.
//
//	Saída:
//	  distanceMeters: distância total em metros (bifurcações são contadas);
//	  features: lista de geojson features;
//	  err: objeto padrão de erro do golang.
func (e *Osm) WayJoinGeoJSonFeatures(id int64, timeout time.Duration) (distanceMeters float64, features string, err error) {
	var idList = make([]int64, 0)
	var idListGarbage = make([]int64, 0)
	var name string
	var found bool
	var coordinates = make([][2]float64, 0)

	var dbWay Way

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = e.ClientWay.FindOne(ctx, bson.M{"_id": id}).Decode(&dbWay)
	if err != nil {
		cancel()
		return
	}
	cancel()

	if dbWay.Id == 0 {
		err = errors.New("way not found")
		return
	}

	if name, found = dbWay.Tag["name"]; !found {
		err = errors.New("property name not found")
		return
	}

	idList = append(idList, dbWay.Id)
	idListGarbage = append(idListGarbage, dbWay.Id)
	coordinates = append(coordinates, dbWay.Loc.Coordinates...)

	var locList = [][2]float64{dbWay.LocFirst, dbWay.LocLast}

	features = dbWay.GeoJSonFeature + ","
	distanceMeters = dbWay.DistanceTotal

	for {
		if len(locList) == 0 {
			break
		}

		dbWay, err = e.wayJoinLast(locList[0], timeout, &idListGarbage, name)
		locList = locList[1:]

		if err != nil {
			continue
		}

		if dbWay.Id == 0 {
			continue
		}

		locList = append(locList, dbWay.LocFirst, dbWay.LocLast)

		features += dbWay.GeoJSonFeature + ","
		distanceMeters += dbWay.DistanceTotal
	}

	features = strings.TrimSuffix(features, ",")
	return
}

func (e *Osm) WayJoinQueryGeoJSonFeatures(query interface{}, timeout time.Duration) (distanceMeters float64, features string, err error) {
	var cursor *mongo.Cursor
	var dbWay Way

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	cursor, err = e.ClientWay.Find(ctx, query)
	if err != nil {
		cancel()
		return
	}
	cancel()

	for cursor.Next(context.Background()) {
		err = cursor.Decode(&dbWay)
		if err != nil {
			return
		}
		features += dbWay.GeoJSonFeature + ","
		distanceMeters += dbWay.DistanceTotal
	}

	features = strings.TrimSuffix(features, ",")
	return
}

//func (e *MongoDbOsm) WayJoin(id int64, timeout time.Duration) (way goosm.Way, err error) {
//	var idList = make([]int64, 0)
//	var idListGarbage = make([]int64, 0)
//	var name string
//	var found bool
//	var coordinates = make([][2]float64, 0)
//	var distance = make([]float64, 4)
//
//	var dbWay Way
//
//	ctx, cancel := context.WithTimeout(context.Background(), timeout)
//	err = e.ClientWay.FindOne(ctx, bson.M{"_id": id}).Decode(&dbWay)
//	if err != nil {
//		cancel()
//		return
//	}
//	cancel()
//
//	if dbWay.Id == 0 {
//		err = errors.New("way not found")
//		return
//	}
//
//	if name, found = dbWay.Tag["name"]; !found {
//		err = errors.New("property name not found")
//		return
//	}
//
//	idList = append(idList, dbWay.Id)
//	idListGarbage = append(idListGarbage, dbWay.Id)
//	coordinates = append(coordinates, dbWay.Loc.Coordinates...)
//
//	var locList = [][2]float64{dbWay.LocFirst, dbWay.LocLast}
//
//	for {
//		if len(locList) == 0 {
//			break
//		}
//
//		dbWay, err = e.wayJoinLast(locList[0], timeout, &idListGarbage, name)
//		locList = locList[1:]
//
//		if err != nil {
//			continue
//		}
//
//		if dbWay.Id == 0 {
//			continue
//		}
//
//		locList = append(locList, dbWay.LocFirst, dbWay.LocLast)
//
//		// O objetivo é deixar ⋁ próximos
//
//		//                                 ⋁            ⋁
//		//wayA.LocFirst wayB.LocFirst - A: 0 1 2 3 - B: 0 1 2 3 - B invertido na frente de A
//		distance[0] = goosm.DistanceBetweenTwoRadPoints([2]float64{util.DegreesToRadians(coordinates[0][goosm.Longitude]), util.DegreesToRadians(coordinates[0][goosm.Latitude])}, dbWay.Rad[0]).Meters
//
//		//                                ⋁                  ⋁
//		//wayA.LocFirst wayB.LocLast - A: 0 1 2 3 - B: 0 1 2 3 - B na frente de A
//		distance[1] = goosm.DistanceBetweenTwoRadPoints([2]float64{util.DegreesToRadians(coordinates[0][goosm.Longitude]), util.DegreesToRadians(coordinates[0][goosm.Latitude])}, dbWay.Rad[len(dbWay.Rad)-1]).Meters
//
//		//                                      ⋁      ⋁
//		//wayA.LocLast wayB.LocFirst - A: 0 1 2 3 - B: 0 1 2 3 - B depois de A
//		distance[2] = goosm.DistanceBetweenTwoRadPoints([2]float64{util.DegreesToRadians(coordinates[len(coordinates)-1][goosm.Longitude]), util.DegreesToRadians(coordinates[len(coordinates)-1][goosm.Latitude])}, dbWay.Rad[0]).Meters
//
//		//                                     ⋁            ⋁
//		//wayA.LocLast wayB.LocLast - A: 0 1 2 3 - B: 0 1 2 3 - B invertido depois de A
//		distance[3] = goosm.DistanceBetweenTwoRadPoints([2]float64{util.DegreesToRadians(coordinates[len(coordinates)-1][goosm.Longitude]), util.DegreesToRadians(coordinates[len(coordinates)-1][goosm.Latitude])}, dbWay.Rad[len(dbWay.Rad)-1]).Meters
//
//		minDistance := math.MaxFloat64
//		minDistanceK := 0
//		for k := range distance {
//			if min := math.Min(distance[k], minDistance); min != minDistance {
//				minDistance = min
//				minDistanceK = k
//			}
//		}
//
//		switch minDistanceK {
//		case 0:
//			tmp := make([][2]float64, len(dbWay.Loc.Coordinates))
//			index := 0
//			for i := len(dbWay.Loc.Coordinates) - 1; i >= 0; i-- { // B invertido na frente de A
//				tmp[index] = dbWay.Loc.Coordinates[i]
//				index++
//			}
//			coordinates = append(tmp, coordinates...)
//
//		case 1:
//			coordinates = append(dbWay.Loc.Coordinates, coordinates...) // B na frente de A
//
//		case 2:
//			coordinates = append(coordinates, dbWay.Loc.Coordinates...) // B depois de A
//
//		case 3:
//			tmp := make([][2]float64, len(dbWay.Loc.Coordinates))
//			index := 0
//			for i := len(dbWay.Loc.Coordinates) - 1; i >= 0; i-- { // coloca B invertido depois de A
//				tmp[index] = dbWay.Loc.Coordinates[i]
//				index++
//			}
//			coordinates = append(coordinates, tmp...)
//		}
//
//		idList = append(idList, dbWay.Id)
//	}
//
//	way = goosm.Way{}
//	way.Id = id
//	way.Tag = map[string]string{"name": name}
//	way.Loc = coordinates
//	err = way.Init()
//	if err != nil {
//		return
//	}
//	way.IdList = idList
//
//	way.MakeGeoJSonFeature() //fixme: apagar
//	err = way.PolygonList(50)
//	if err != nil {
//		return
//	}
//
//	return
//}

func (e *Osm) GetWayById(id int64) (way goosm.Way, err error) {
	var tmpWay Way

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	err = e.ClientWay.FindOne(ctx, bson.M{"_id": id}).Decode(&tmpWay)
	if err != nil {
		return
	}

	way = tmpWay.Way()
	return
}

func (e *Osm) GetWay(id int64, timeout time.Duration) (way goosm.Way, err error) {
	var tmpWay Way

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = e.ClientWay.FindOne(ctx, bson.M{"_id": id}).Decode(&tmpWay)
	if err != nil {
		return
	}

	way = tmpWay.Way()
	return
}

func (e *Osm) SetWay(wayList *[]goosm.Way, timeout time.Duration) (err error) {

	var tmpWay Way
	var listToInsert = make([]interface{}, len(*wayList))
	var listToDelete = make([]int64, len(*wayList))
	for key, way := range *wayList {
		listToDelete[key] = way.Id
		listToInsert[key] = tmpWay.ToDbWay(way)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_, _ = e.ClientWay.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": listToDelete}})
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	_, err = e.ClientWay.InsertMany(ctx, listToInsert)
	cancel()

	return
}

func (e *Osm) createTableNode() (err error) {
	e.ClientNode = e.Client.Database(constants.KMongoDBDatabase).Collection(constants.KMongoDBCollectionNode)

	indexes := e.ClientNode.Indexes()

	var cursor *mongo.Cursor
	cursor, err = indexes.List(context.Background())
	if err != nil {
		return
	}

	results := make([]bson.M, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return
	}

	pass := false
	for _, result := range results {
		if result["name"] == "__loc__" {
			pass = true
			break
		}
	}

	if !pass {
		name := "__loc__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"loc": "2dsphere"},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__tags__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"tag": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}
	}

	return
}

func (e *Osm) createTableWrongWay() (err error) {
	e.ClientWrongWay = e.Client.Database(constants.KMongoDBDatabase).Collection(constants.KMongoDBCollectionWrongWay)

	indexes := e.ClientWrongWay.Indexes()

	var cursor *mongo.Cursor
	cursor, err = indexes.List(context.Background())
	if err != nil {
		return
	}

	results := make([]bson.M, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return
	}

	pass := false
	for _, result := range results {
		if result["name"] == "__loc__" {
			pass = true
			break
		}
	}

	if !pass {
		name := "__loc__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"loc": "2dsphere"},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__tags__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"tag": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}
	}

	return
}

func (e *Osm) createTableWay() (err error) {
	e.ClientWay = e.Client.Database(constants.KMongoDBDatabase).Collection(constants.KMongoDBCollectionWay)

	indexes := e.ClientWay.Indexes()

	var cursor *mongo.Cursor
	cursor, err = indexes.List(context.Background())
	if err != nil {
		return
	}

	results := make([]bson.M, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return
	}

	pass := false
	for _, result := range results {
		if result["name"] == "__tags__" {
			pass = true
			break
		}
	}

	if !pass {
		name := "__loc__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"loc": "2dsphere"},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__tags__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"tag": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__locFirst__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"locFirst": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__locLast__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"locLast": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}

		name = "__idList__"
		_, err = indexes.CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys: bson.M{"idList": 1},
				Options: &options.IndexOptions{
					Name: &name,
				},
			},
		)
		if err != nil {
			return
		}
	}

	return
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
