package main

import (
	"fmt"
	"goosm/compress"
	"goosm/goosm"
	downloadApiV06 "goosm/goosm/download"
	"goosm/plugin/mongodb"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	var err error
	var done = make(chan struct{})
	var timeout = 10 * time.Second
	var terminalInterval = 2000 * time.Millisecond
	var fileDownloadName = "http://download.geofabrik.de/south-america/brazil/sul-latest.osm.pbf"
	var fileSaveName = "../../planet-221010.osm.1.pbf" //"./sul-latest.osm.pbf"
	var fileTmpName = "./node.sul.tmp"

	fmt.Println("Starting file download. This may take a while. It's ~300MB.")

	// English: Download the binary file with the map from Open Street Maps
	// Português: Faz o download do arquivo binário com o mapa do Open Street Maps
	err = downloadGeoFabrikMap(
		fileDownloadName,
		fileSaveName,
		terminalInterval,
	)
	if err != nil {
		panic(err)
	}

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
	err = compressData.Open(fileTmpName)
	if err != nil {
		panic(err)
	}
	defer compressData.Close()

	start := time.Now()

	// English: Process Open Street Maps binary file
	// Português: Processa o arquivo binário do Open Street Maps
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

	// English: Defines the binary search interface to process the Open Street Maps file
	// Português: Define a interface da busca binária para processar o arquivo do Open Street Maps
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
	dbNode := &mongodb.DbNode{}
	_, err = dbNode.New(conn, database, "node", timeout)
	if err != nil {
		err = fmt.Errorf("setupDatabase: the function dbNode.New() returned an error: %v", err)
		return
	}

	// English: Defines the object database for nodes
	// Português: Define o objeto de banco de dados para nodes
	dbWay := &mongodb.DbWay{}
	_, err = dbWay.New(conn, database, "way", timeout)
	if err != nil {
		err = fmt.Errorf("setupDatabase: the function dbWay.New() returned an error: %v", err)
		return
	}

	return
}

// downloadGeoFabrikMap
//
// English:
//
// # Download a small map file from the Geo FabrikMap website
//
// Português:
//
// Faz o download de um arquivo de mapa pequeno do site Geo FabrikMap
func downloadGeoFabrikMap(downloadPath, fileToSavePath string, terminalInterval time.Duration) (err error) {
	// English: If the file exists, do nothing
	// Português: Se o arquivo existe, não faz nada
	if _, err = os.Stat(fileToSavePath); err == nil {
		return
	}

	var bytesDownloaded int64
	var fileToSave *os.File
	var fileToDownload *http.Response
	var done = make(chan struct{})

	fileToSave, err = os.Create(fileToSavePath)
	if err != nil {
		err = fmt.Errorf("error trying to create the file '%v': %v", fileToSavePath, err)
		return err
	}
	defer func() {
		_ = fileToSave.Close()
	}()

	fileToDownload, err = http.Get(downloadPath)
	if err != nil {
		err = fmt.Errorf("error trying to download the file '%v': %v", downloadPath, err)
		return err
	}
	defer func() {
		_ = fileToDownload.Body.Close()
	}()

	parcialReportTicker := time.NewTicker(terminalInterval)
	go func() {
		for {
			select {
			case <-parcialReportTicker.C:
				var info fs.FileInfo
				info, err = fileToSave.Stat()
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("Download\n")
				log.Printf("bytes downloaded: %vMB\n\n", info.Size()/1024/1024)

			case <-done:

				parcialReportTicker.Stop()
				return
			}
		}
	}()

	bytesDownloaded, err = io.Copy(fileToSave, fileToDownload.Body)
	if err != nil {
		err = fmt.Errorf("error trying to save the file '%v': %v", fileToSavePath, err)
		return err
	}
	done <- struct{}{}

	fmt.Printf("File downloaded. Total: %vMB\n", bytesDownloaded/1024/1024)
	return
}
