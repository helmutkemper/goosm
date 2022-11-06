package main

import (
	"fmt"
	"goosm/goosm"
	"goosm/plugin/mongodb"
	"time"
)

func main() {
	var err error
	var timeout = 10 * time.Second

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	dbNode := &mongodb.DbNode{}
	_, err = dbNode.New("mongodb://127.0.0.1:27016/", "osm", "node", timeout)
	if err != nil {
		panic(err)
	}

	var node goosm.Node
	start := time.Now()
	node, err = dbNode.GetById(273316)
	if err != nil {
		panic(err)
	}
	duration := time.Since(start)
	fmt.Printf("duração da requsição: %v\n", duration)

	// English: to see this information on a map, just go to https://geojson.io/ and paste it inside the square brackets
	//   in "features": [], on the line 3
	// Português: para vê esta informação em um mapa, basta entrar no site https://geojson.io/ e colar ela dentro dos
	//   couchettes em "features": [], na linha 3
	fmt.Printf("Node:\n%v\n", node.GeoJSonFeature)

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	dbWay := &mongodb.DbWay{}
	_, err = dbWay.New("mongodb://127.0.0.1:27016/", "osm", "way", timeout)
	if err != nil {
		panic(err)
	}

	var way goosm.Way
	start = time.Now()
	way, err = dbWay.GetById(10492763)
	if err != nil {
		panic(err)
	}
	duration = time.Since(start)
	fmt.Printf("duração da requsição: %v\n", duration)

	// English: to see this information on a map, just go to https://geojson.io/ and paste it inside the square brackets
	//   in "features": [], on the line 3
	// Português: para vê esta informação em um mapa, basta entrar no site https://geojson.io/ e colar ela dentro dos
	//   couchettes em "features": [], na linha 3
	fmt.Printf("Way:\n%v\n", way.GeoJSonFeature)
}
