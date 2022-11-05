package main

import (
	"fmt"
	datasource "goosm/businessRules/dataSource"
	"goosm/goosm"
)

func main() {
	var err error
	err = datasource.Linker.Init(datasource.KMongoDB)
	if err != nil {
		panic(err)
	}

	var node goosm.Node
	node, err = datasource.Linker.Osm.GetNodeById(90200063)
	if err != nil {
		panic(err)
	}

	// English: to see this information on a map, just go to https://geojson.io/ and paste it inside the square brackets
	//   in "features": [], on the line 3
	// Português: para vê esta informação em um mapa, basta entrar no site https://geojson.io/ e colar ela dentro dos
	//   couchettes em "features": [], na linha 3
	fmt.Printf("\n\n%v\n", node.GeoJSonFeature)
}
