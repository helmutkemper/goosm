package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"goosm/goosm"
	"time"
)

type DbWay struct {
	timeout    time.Duration
	Client     *mongo.Client
	Collection *mongo.Collection
}

// SetTimeout
//
// English:
//
// Determines timeout for all functions
//
//	Input:
//	  timeout: maximum time for operation
//
// Português:
//
// Determina o timeout para todas as funções
//
//	Entrada:
//	  timeout: tempo máximo para a operação
func (e *DbWay) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// Connect
//
// English:
//
// Connect to the database
//
//	Input:
//	  connection: database connection string. eg. "mongodb://127.0.0.1:27016/"
//	  args: maintained by interface compatibility
//
// Português:
//
// Conecta ao banco de dados
//
//	Entrada:
//	  connection: string de conexão ao banco de dados. Ex: "mongodb://127.0.0.1:27016/"
//	  args: mantido por compatibilidade da interface
func (e *DbWay) Connect(connection string, _ ...interface{}) (err error) {
	e.Client, err = mongo.NewClient(options.Client().ApplyURI(connection))
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.Connect().NewClient().error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Connect(ctx)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.Connect().Connect().error: %v", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Ping(ctx, readpref.Primary())
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.Connect().Ping().error: %v", err)
		return
	}
	return
}

// Close
//
// English:
//
// # Close the connection to the database
//
// Português:
//
// Fecha a conexão com o banco de dados
func (e *DbWay) Close() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Disconnect(ctx)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.Close().Disconnect().error: %v", err)
		return
	}
	return
}

// New
//
// English:
//
// Prepare the database for use
//
//	Input:
//	  connection: database connection string. Eg: "mongodb://127.0.0.1:27016/"
//	  database: database name. Eg. "osm"
//	  collection: collection name within the database. Eg. "way"
//
//	Output:
//	  referenceInitialized: database way object ready to use
//	  err: golang error object
//
// Português:
//
// Prepara o banco de dados para uso
//
//	Entrada:
//	  connection: string de conexão ao banco de dados. Ex: "mongodb://127.0.0.1:27016/"
//	  database: nome do banco de dados. Ex: "osm"
//	  collection: nome da coleção dentro do banco de dados. Ex: "way"
//
//	Saída:
//	  referenceInitialized: objeto do banco de dados pronto para uso
//	  err: objeto golang error
func (e *DbWay) New(connection, database, collection string, timeout time.Duration) (referenceInitialized interface{}, err error) {
	e.SetTimeout(timeout)

	if err = e.Connect(connection); err != nil {
		return
	}
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.New().Connect().error: %v", err)
		return
	}

	if err = e.createTable(database, collection); err != nil {
		return
	}
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.New().createTable().error: %v", err)
		return
	}

	return e, err
}

// SetOne
//
// English:
//
// Insert a single way into the database
//
//	Input:
//	  way: reference to object goosm.Way
//
// Português:
//
// Insere um único way no banco de dados
//
//	Entrada:
//	  way: referencia ao objeto goosm.Way.
func (e *DbWay) SetOne(way *goosm.Way) (err error) {
	var wayDb = Way{}
	wayDb.ToDbWay(way)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.Collection.InsertOne(ctx, wayDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.SetOne().InsertOne().error: %v", err)
		return
	}
	return
}

// GetById
//
// English:
//
// Returns a way according to ID
//
//	Input:
//	  id: ID in the Open Street Maps project pattern
//
// Português:
//
// Retorna um way de acordo com o ID
//
//	Entrada:
//	  id: ID no padrão do projeto Open Street Maps.
func (e *DbWay) GetById(id int64) (way goosm.Way, err error) {
	var wayDb Way
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&wayDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.GetById().FindOne().error: %v", err)
		return
	}

	way = wayDb.ToOsmWay()
	return
}

// SetMany
//
// English:
//
// Insert a block of ways into the database
//
//	Input:
//	  list: reference to slice with []goosm.Way objects
//
// Português:
//
// Insere um bloco de ways no banco de dados
//
//	Entrada:
//	  list: referência ao slice com os objetos []goosm.Way
func (e *DbWay) SetMany(list *[]goosm.Way) (err error) {
	wayDb := Way{}
	var listDb = make([]interface{}, len(*list))
	for key, way := range *list {
		wayDb.ToDbWay(&way)
		listDb[key] = wayDb
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.Collection.InsertMany(ctx, listDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.SetMany().InsertMany().error: %v", err)
		return
	}
	return
}

// createTable
//
// English:
//
// Create the collection and indexes
//
//	Input:
//	  database: database name. Eg. "osm"
//	  collection: collection name within the database. Eg. "way"
//
// Português:
//
// Cria a coleção e os índices
//
//	Entrada:
//	  database: nome do banco de dados. Ex: "osm"
//	  collection: nome da coleção dentro do banco de dados. Ex: "way"
func (e *DbWay) createTable(database, collection string) (err error) {
	e.Collection = e.Client.Database(database).Collection(collection)

	indexes := e.Collection.Indexes()

	var cursor *mongo.Cursor
	cursor, err = indexes.List(context.Background())
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.createTable().List().error: %v", err)
		return
	}

	results := make([]bson.M, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		err = fmt.Errorf("mongodb.DbWay.createTable().All().error: %v", err)
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
			err = fmt.Errorf("mongodb.DbWay.createTable().CreateOne(loc).error: %v", err)
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
			err = fmt.Errorf("mongodb.DbWay.createTable().CreateOne(tag).error: %v", err)
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
			err = fmt.Errorf("mongodb.DbWay.createTable().CreateOne(locFirst).error: %v", err)
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
			err = fmt.Errorf("mongodb.DbWay.createTable().CreateOne(locLast).error: %v", err)
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
			err = fmt.Errorf("mongodb.DbWay.createTable().CreateOne(idList).error: %v", err)
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
