package main

import (
	"fmt"
	datasource "goosm/businessRules/dataSource"
	"goosm/compress"
	"goosm/goosm"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

// 22:27
// Tempo total: 33m0.474874167s

func main() {

	var err error
	var done = make(chan struct{})
	var terminalInterval = 2000 * time.Millisecond

	fmt.Println("Starting file download. This may take a while. It's ~300MB.")

	err = downloadGeoFabrikMap(terminalInterval)
	if err != nil {
		panic(err)
	}

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

	start := time.Now()
	var binarySearch = &goosm.PbfProcess{}

	// English: determines the database interface
	// Português: determina a interface de banco de dados
	binarySearch.SetDatabase(datasource.Linker.Osm)

	// English: determines the maximum database response time to do 100 simultaneous inserts
	// Português: determina o tempo máximo de resposta do banco de dados para fazer 100 inserções simultâneas
	binarySearch.SetDatabaseTimeout(10 * 60 * time.Second)

	parcialReportTicker := time.NewTicker(terminalInterval)
	go func() {
		firstReport := make([]uint64, 0)
		lastReport := make([]uint64, 0)
		lastResult := uint64(0)
		for {
			select {
			case <-parcialReportTicker.C:
				nodes := uint64(0)
				nodes, _ = binarySearch.GetPartialNumberOfProcessedData()

				if len(firstReport) < 10 {
					firstReport = append(firstReport, nodes-lastResult)
				}

				lastReport = append(lastReport, nodes-lastResult)
				if len(lastReport) > 10 {
					lastReport = lastReport[:10]
				}

				firstInterval := uint64(0)
				for k := range firstReport {
					firstInterval += firstReport[k]
				}
				firstInterval = firstInterval / uint64(len(firstReport))

				lastInterval := uint64(0)
				for k := range lastReport {
					lastInterval += lastReport[k]
				}
				lastInterval = lastInterval / uint64(len(lastReport))

				log.Println("Partial report:")
				log.Printf("nodes: %v\n", nodes)
				log.Printf("diference: %2.2f", float64(lastInterval/firstInterval*100))

				lastResult = nodes

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
	err = binarySearch.WrongWayParser("./sul-latest.osm.pbf", "./insertionTime.csv")
	log.Printf("time total: %v", time.Since(start))
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
	// English: If the file exists, do nothing
	// Português: Se o arquivo existe, não faz nada
	if _, err = os.Stat("./sul-latest.osm.pbf"); err == nil {
		return
	}

	var bytesDownloaded int64
	var fileToSave *os.File
	var fileToDownload *http.Response
	var done = make(chan struct{})

	fileToSave, err = os.Create("./sul-latest.osm.pbf")
	if err != nil {
		err = fmt.Errorf("error trying to create the file './sul-latest.osm.pbf': %v", err)
		return err
	}
	defer func() {
		_ = fileToSave.Close()
	}()

	fileToDownload, err = http.Get("http://download.geofabrik.de/south-america/brazil/sul-latest.osm.pbf")
	if err != nil {
		err = fmt.Errorf("error trying to download the file 'http://download.geofabrik.de/south-america/brazil/sul-latest.osm.pbf': %v", err)
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
		err = fmt.Errorf("error trying to save the file './sul-latest.osm.pbf': %v", err)
		return err
	}
	done <- struct{}{}

	fmt.Printf("File downloaded. Total: %vMB\n", bytesDownloaded/1024/1024)
	return
}
