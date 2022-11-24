package main

import (
	"fmt"
	"goosm/compress"
	"goosm/goosm"
	downloadApiV06 "goosm/goosm/download"
	"goosm/plugin/mongodb"
	"log"
	"time"
)

func main() {

	var err error
	var done = make(chan struct{})
	var timeout = 10 * time.Second
	var terminalInterval = 2000 * time.Millisecond
	var fileSaveName = "../commonFiles/sul-latest.osm.pbf"
	var fileTmpName = "../commonFiles/sul-latest.tmp"

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	var dbNode goosm.InterfaceDbNode

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	var dbWay goosm.InterfaceDbWay

	// English: Make the database connection
	// Português: Faz a conexão do banco de dados
	dbWay, dbNode, err = setupDatabase("mongodb://127.0.0.1:27016/", "osm", timeout)
	if err != nil {
		panic(err)
	}

	// English: Defines the object responsible for downloading, when information is not found in the binary file
	// Português: Define o objeto responsável por download, quando uma informação não é encontrada no arquivo binário
	download := &downloadApiV06.DownloadApiV06{}

	// English: Compress and process the file '.sul-latest.osm.pbf' using a binary search in memory and file to save processing time
	// Português: Comprime e processa o arquivo './sul-latest.osm.pbf' usando uma busca binária em memória e arquivo para ganhar tempo de processamento
	compressData := &compress.Compress{}
	compressData.Init(100)
	err = compressData.OpenForSearch(fileTmpName)
	if err != nil {
		panic(err)
	}
	defer compressData.Close()

	start := time.Now()

	// English: Process Create Street Maps binary file
	// Português: Processa o arquivo binário do Create Street Maps
	var osmFileProcess = &goosm.PbfProcess{}

	// English: determines the database interface for node
	// Português: determina a interface de banco de dados para node
	osmFileProcess.SetDatabaseNode(dbNode)

	// English: determines the database interface for way
	// Português: determina a interface de banco de dados para way
	osmFileProcess.SetDatabaseWay(dbWay)

	// English: determines the download interface for when a point is not found in the file
	// Português: determina a interface de download para quando um ponto não é encontrado no arquivo
	osmFileProcess.SetDownloadApi(download)

	// English: determines the maximum database response time to do 100 simultaneous inserts
	// Português: determina o tempo máximo de resposta do banco de dados para fazer 100 inserções simultâneas
	osmFileProcess.SetDatabaseTimeout(10 * 60 * time.Second)

	// English: Defines the binary search interface to process the Create Street Maps file
	// Português: Define a interface da busca binária para processar o arquivo do Create Street Maps
	osmFileProcess.SetCompress(compressData)

	parcialReportTicker := time.NewTicker(terminalInterval)
	go func() {
		for {
			select {
			case <-parcialReportTicker.C:
				nodes := uint64(0)
				ways := uint64(0)
				nodes, ways = osmFileProcess.GetPartialNumberOfProcessedData()
				log.Println("Partial report:")
				log.Printf("nodes: %v\n", nodes)
				log.Printf("ways: %v\n\n", ways)

			case <-done:
				parcialReportTicker.Stop()
				return
			}
		}
	}()

	// English: process the file. although the unique responsibility is three functions, binary search, database for nodes
	//   and database for ways, 7.9 trillion of points greatly increases the computational cost.
	// Português: processa o arquivo. embora a responsabilidade única peça que sejam três funções, busca binária, banco de
	//   dados para nodes e banco de dados para ways, 7.9 trilhões de pontos elevam muito o custo computacional.
	_, _, err = osmFileProcess.CompleteParser(fileSaveName)
	log.Printf("time total: %v", time.Since(start))
	if err != nil {
		panic(err)
	}

	done <- struct{}{}

}

// setupDatabase
//
// English:
//
// Make MongoDB database connection
//
//	Input:
//	  conn: connection string. eg. "mongodb://127.0.0.1:27016/"
//	  database: database name
//	  timeout: maximum operation time
//
//	Output:
//	  way: object compatible with the goosm.InterfaceDbWay interface
//	  node: object compatible with the goosm.InterfaceDbNode interface
//	  err: golang error object
//
// Português:
//
// Faz a conexão do banco de dados MongoDB
//
//	Entrada:
//	  conn: string de conexão. ex,: "mongodb://127.0.0.1:27016/"
//	  database: database name
//	  timeout: tempo máximo da operação
//
//	Saída:
//	  way: objeto compatível com a interface goosm.InterfaceDbWay
//	  node: objeto compatível com a interface goosm.InterfaceDbNode
//	  err: objeto de erro golang
func setupDatabase(conn, database string, timeout time.Duration) (way *mongodb.DbWay, node *mongodb.DbNode, err error) {
	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	node = &mongodb.DbNode{}
	_, err = node.New(conn, database, "node", timeout)
	if err != nil {
		err = fmt.Errorf("setupDatabase: the function dbNode.New() returned an error: %v", err)
		return
	}

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	way = &mongodb.DbWay{}
	_, err = way.New(conn, database, "way", timeout)
	if err != nil {
		err = fmt.Errorf("setupDatabase: the function dbWay.New() returned an error: %v", err)
		return
	}

	return
}
