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

type DbNode struct { //nolint:typecheck
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
func (e *DbNode) SetTimeout(timeout time.Duration) {
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
func (e *DbNode) Connect(connection string, _ ...interface{}) (err error) {
	e.Client, err = mongo.NewClient(options.Client().ApplyURI(connection))
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.Connect().NewClient().error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Connect(ctx)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.Connect().Connect().error: %v", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Ping(ctx, readpref.Primary())
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.Connect().Ping().error: %v", err)
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
func (e *DbNode) Close() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Client.Disconnect(ctx)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.Close().Disconnect().error: %v", err)
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
//	  collection: collection name within the database. Eg. "node"
//
//	Output:
//	  referenceInitialized: database node object ready to use
//	  err: golang error object
//
// Português:
//
// Prepara o banco de dados para uso
//
//	Entrada:
//	  connection: string de conexão ao banco de dados. Ex: "mongodb://127.0.0.1:27016/"
//	  database: nome do banco de dados. Ex: "osm"
//	  collection: nome da coleção dentro do banco de dados. Ex: "node"
//
//	Saída:
//	  referenceInitialized: objeto do banco de dados pronto para uso
//	  err: objeto golang error
func (e *DbNode) New(connection, database, collection string) (referenceInitialized interface{}, err error) { //nolint:typecheck
	if err = e.Connect(connection); err != nil {
		return
	}
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.New().Connect().error: %v", err)
		return
	}

	if err = e.createTable(database, collection); err != nil {
		return
	}
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.New().createTable().error: %v", err)
		return
	}

	return e, err
}

// SetOne
//
// English:
//
// Insert a single node into the database
//
//	Input:
//	  node: reference to object goosm.Node
//
// Português:
//
// Insere um único node no banco de dados
//
//	Entrada:
//	  node: referencia ao objeto goosm.Node
func (e *DbNode) SetOne(node *goosm.Node) (err error) { //nolint:typecheck
	var nodeDb = Node{}
	nodeDb.ToDbNode(node)

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.Collection.InsertOne(ctx, nodeDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.SetOne().InsertOne().error: %v", err)
		return
	}
	return
}

// GetById
//
// English:
//
// Returns a node according to ID
//
//	Input:
//	  id: ID in the Open Street Maps project pattern
//
// Português:
//
// Retorna um node de acordo com o ID
//
//	Entrada:
//	  id: ID no padrão do projeto Open Street Maps
func (e *DbNode) GetById(id int64) (node goosm.Node, err error) { //nolint:typecheck
	var nodeDb Node
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	err = e.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&nodeDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.GetById().FindOne().error: %v", err)
		return
	}

	node = nodeDb.ToOsmNode()
	return
}

// SetMany
//
// English:
//
// Insert a block of nodes into the database
//
//	Input:
//	  list: reference to slice with []goosm.Node objects
//
// Português:
//
// Insere um bloco de nodes no banco de dados
//
//	Entrada:
//	  list: referência ao slice com os objetos []goosm.Node
func (e *DbNode) SetMany(list *[]goosm.Node) (err error) { //nolint:typecheck
	nodeDb := Node{}
	var listDb = make([]interface{}, len(*list))
	for key, node := range *list {
		nodeDb.ToDbNode(&node)
		listDb[key] = nodeDb
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err = e.Collection.InsertMany(ctx, listDb)
	cancel()
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.SetMany().InsertMany().error: %v", err)
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
//	  collection: collection name within the database. Eg. "node"
//
// Português:
//
// Cria a coleção e os índices
//
//	Entrada:
//	  database: nome do banco de dados. Ex: "osm"
//	  collection: nome da coleção dentro do banco de dados. Ex: "node"
func (e *DbNode) createTable(database, collection string) (err error) {
	e.Collection = e.Client.Database(database).Collection(collection)

	indexes := e.Collection.Indexes()

	var cursor *mongo.Cursor
	cursor, err = indexes.List(context.Background())
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.createTable().List().error: %v", err)
		return
	}

	results := make([]bson.M, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		err = fmt.Errorf("mongodb.DbNode.createTable().All().error: %v", err)
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
			err = fmt.Errorf("mongodb.DbNode.createTable().CreateOne().error: %v", err)
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
			err = fmt.Errorf("mongodb.DbNode.createTable().CreateOne().error: %v", err)
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
