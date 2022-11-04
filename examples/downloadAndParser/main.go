package main

import (
	"fmt"
	datasource "goosm/businessRules/dataSource"
	"goosm/compress"
	"goosm/goosm"
	downloadApiV06 "goosm/goosm/download"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	var err error
	var done = make(chan struct{})
	var terminalInterval = 2000 * time.Millisecond

	fmt.Println("Starting file download. This may take a while. It's ~300MB.")

	err = downloadGeoFabrikMap(terminalInterval)

	err = datasource.Linker.Init(datasource.KMongoDB)
	if err != nil {
		panic(err)
	}

	// English: Compress and process the file '.sul-latest.osm.pbf' using a binary search in memory and file to save processing time
	// Português: Comprime e processa o arquivo './sul-latest.osm.pbf' usando uma busca binária em memória e arquivo para ganhar tempo de processamento
	compressData := &compress.Compress{}
	compressData.Init(100)
	err = compressData.Open("./node.sul.tmp")
	if err != nil {
		panic(err)
	}
	defer compressData.Close()

	totalNodes := uint64(0)
	totalWays := uint64(0)
	start := time.Now()
	var binarySearch = &goosm.PbfProcess{}

	// English: determines the database interface
	// Português: determina a interface de banco de dados
	binarySearch.SetDatabase(datasource.Linker.Osm)

	// English: determines the download interface for when a point is not found in the file
	// Português: determina a interface de download para quando um ponto não é encontrado no arquivo
	binarySearch.SetDownloadApi(&downloadApiV06.DownloadApiV06{})

	// English: determines the maximum database response time to do 100 simultaneous inserts
	// Português: determina o tempo máximo de resposta do banco de dados para fazer 100 inserções simultâneas
	binarySearch.SetDatabaseTimeout(10 * 60 * time.Second)
	binarySearch.SetCompress(compressData)

	parcialReportTicker := time.NewTicker(terminalInterval)
	go func() {
		for {
			select {
			case <-parcialReportTicker.C:
				nodes := uint64(0)
				ways := uint64(0)
				nodes, ways = binarySearch.GetPartialNumberOfProcessedData()
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
	totalNodes, totalWays, err = binarySearch.CompleteParser("./sul-latest.osm.pbf")
	log.Printf("time total: %v", time.Since(start))
	log.Printf("nodes: %v", totalNodes)
	log.Printf("ways: %v", totalWays)
	if err != nil {
		panic(err)
	}

	done <- struct{}{}

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
func downloadGeoFabrikMap(terminalInterval time.Duration) (err error) {
	var bytesDownloaded int64
	var fileToSave *os.File
	var fileToDownload *http.Response
	var done = make(chan struct{})

	fileToSave, err = os.Create("./sul-latest.osm.pbf")
	if err != nil {
		err = fmt.Errorf("error trying to create the file './sul-latest.osm.pbf': %v", err)
		panic(err)
	}
	defer func() {
		_ = fileToSave.Close()
	}()

	fileToDownload, err = http.Get("http://download.geofabrik.de/south-america/brazil/sul-latest.osm.pbf")
	if err != nil {
		err = fmt.Errorf("error trying to download the file 'http://download.geofabrik.de/south-america/brazil/sul-latest.osm.pbf': %v", err)
		panic(err)
	}
	defer func() {
		_ = fileToDownload.Body.Close()
	}()

	parcialReportTicker := time.NewTicker(terminalInterval)
	go func() {
		for {
			select {
			case <-parcialReportTicker.C:
				fi, err := fileToSave.Stat()
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("Download\n")
				log.Printf("bytes downloaded: %vMB\n\n", fi.Size()/1024/1024)

			case <-done:

				parcialReportTicker.Stop()
				return
			}
		}
	}()

	bytesDownloaded, err = io.Copy(fileToSave, fileToDownload.Body)
	if err != nil {
		err = fmt.Errorf("error trying to save the file './sul-latest.osm.pbf': %v", err)
		panic(err)
	}
	done <- struct{}{}

	fmt.Printf("File downloaded. Total: %vMB\n", bytesDownloaded/1024/1024)
	return
}
