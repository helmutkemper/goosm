# goosm

Fast open street maps to database

> Under development

# English

This code shows how to decrease the time of importing `Open Street Maps` maps using binary search to optimize the search 
for data.

### List of examples

| Example                    | Description                                                                                          |
|----------------------------|------------------------------------------------------------------------------------------------------|
| MongoDB Install            | Install the `MongoDB` database on port `27016` and use the `docker/mongodb/data` folder for the data |
| Download and parser        | Download a map and insert the contents into the `MongoDB` database                                   |
| Download and Parser Binary | Allows you to generate a new binary file quickly                                                     |
| Find one by id             | Shows how to use `MongoDB` driver and capture `GeoJson` from geographic information                  |
| Binary search              | Bench mark and example of using binary search                                                        |

### Interfaces

| Name                 | Description                                       |
|----------------------|---------------------------------------------------|
| CompressInterface    | Data compression for binary search                |
| InterfaceDownloadOsm | Download data using Open Street Maps API V0.6     |
| InterfaceConnect     | Database connection, used by node and way objects |
| InterfaceDbNode      | Inserting nodes into the database                 |
| InterfaceDbWay       | Inserting ways into the database                  |

# Português

Este código mostra como diminuir o tempo de importação dos mapas do `Open Street Maps` usando busca binária para 
otimizar a procura por dados

### Lista de exemplos

| Exemplo                    | Descrição                                                                                             |
|----------------------------|-------------------------------------------------------------------------------------------------------|
| MongoDB Install            | Instala o banco de dados `MongoDB` na porta `27016` e usa a pasta `docker/mongodb/data` para os dados |
| Download and parser        | Faz o download de um mapa e insere o conteúdo no banco de dados `MongoDB`                             |
| Download and Parser Binary | Permite gerar um novo arquivo binário de forma rápida.                                                |
| Find one by id             | Mostra como usar o driver `MongoDB` e capturar o `GeoJson` da informação geográfica                   |
| Binary search              | Bench mark e exemplo de uso da busca binária                                                          |

### Interfaces

| Nome                 | Descrição                                                      |
|----------------------|----------------------------------------------------------------|
| CompressInterface    | Compressão de dados para busca binária                         |
| InterfaceDownloadOsm | Faz o download de dados usando a API V0.6 do Opens Street Maps |
| InterfaceConnect     | Conexão do banco de dados, usada pelos objetos node e way      |
| InterfaceDbNode      | Inserção de nodes no banco de dados                            |
| InterfaceDbWay       | Inserção de ways no banco de dados                             |

## Install MongoDB

```golang
package main

import (
	"fmt"
	dockerBuilder "github.com/helmutkemper/iotmaker.docker.builder"
	dockerBuilderNetwork "github.com/helmutkemper/iotmaker.docker.builder.network"
	"io/fs"
	"log"
	"os"
	"time"
)

func main() {
	var err error

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// English: Defines the directory where the golang code runs
	// Português: Define o diretório onde o código golang roda
	err = os.Chdir("../../")
	if err != nil {
		panic(err)
	}

	// English: Create the MongoDB installation folder
	// Português: Cria a pasta de instalação do MongoDB
	_ = os.MkdirAll("./docker/mongodb/data/db", fs.ModePerm)

	// English: Install MongoDB on docker
	// Português: Instala o MongoDB no docker
	err = DockerSupport()
	if err != nil {
		panic(err)
	}
}

// DockerSupport
//
// English:
//
// Removes docker elements that may be left over from previous tests. Any docker element with the term "delete" in the name;
// Install a docker network at address 10.0.0.1 on the computer's network connector;
// Download and install mongodb:latest in docker.
//
//	Note:
//	  * The mongodb:latest image is not removed at the end of the tests so that they can be repeated more easily
//
// Português:
//
// Remove elementos docker que possam ter ficados de testes anteriores. Qualquer elemento docker com o termo "delete" no nome;
// Instala uma rede docker no endereço 10.0.0.1 no conector de rede do computador;
// Baixa e instala o mongodb:latest no docker
//
//	Nota:
//	  * A imagem mongodb:latest não é removida ao final dos testes para que os mesmos passam ser repetidos de forma mais fácil
func DockerSupport() (err error) {

	// English: Docker network controller object
	// Português: Objeto controlador de rede docker
	var netDocker *dockerBuilderNetwork.ContainerBuilderNetwork

	// English: Remove residual docker elements from previous tests (network, images, containers with the term `delete` in the name)
	// Português: Remove elementos docker residuais de testes anteriores (Rede, imagens, containers com o termo `delete` no nome)
	dockerBuilder.SaGarbageCollector()

	// English: Create a docker network (as the gateway is 10.0.0.1, the first address will be 10.0.0.2)
	// Português: Cria uma rede docker (como o gateway é 10.0.0.1, o primeiro endereço será 10.0.0.2)
	netDocker, err = dockerTestNetworkCreate()
	if err != nil {
		return
	}

	// English: Install mongodb on docker
	// Português: Instala o mongodb no docker
	var docker = new(dockerBuilder.ContainerBuilder)
	err = dockerMongoDB(netDocker, docker)
	return
}

// dockerTestNetworkCreate
//
// English:
//
// Create a docker network for the simulations.
//
//	Output:
//	  netDocker: Pointer to the docker network manager object
//	  err: golang error
//
// Português:
//
// Cria uma rede docker para as simulações.
//
//	Saída:
//	  netDocker: Ponteiro para o objeto gerenciador de rede docker
//	  err: golang error
func dockerTestNetworkCreate() (
	netDocker *dockerBuilderNetwork.ContainerBuilderNetwork,
	err error,
) {

	// English: Create a network orchestrator for the container [optional]
	// Português: Cria um orquestrador de rede para o container [opcional]
	netDocker = &dockerBuilderNetwork.ContainerBuilderNetwork{}
	err = netDocker.Init()
	if err != nil {
		err = fmt.Errorf("dockerTestNetworkCreate().error: the function netDocker.Init() returned an error: %v", err)
		return
	}

	// English: Create a network named "cache_delete_after_test"
	// Português: Cria uma rede de nome "cache_delete_after_test"
	err = netDocker.NetworkCreate(
		"cache_delete_after_test",
		"10.0.0.0/16",
		"10.0.0.1",
	)
	if err != nil {
		err = fmt.Errorf("dockerTestNetworkCreate().error: the function netDocker.NetworkCreate() returned an error: %v", err)
		return
	}

	return
}

func dockerMongoDB(
	netDocker *dockerBuilderNetwork.ContainerBuilderNetwork,
	dockerContainer *dockerBuilder.ContainerBuilder,
) (
	err error,
) {

	// English: set a docker network
	// Português: define a rede docker
	dockerContainer.SetNetworkDocker(netDocker)

	// English: define o nome da imagem a ser baixada e instalada.
	// Português: sets the name of the image to be downloaded and installed.
	dockerContainer.SetImageName("mongo:latest")

	// English: defines the name of the MongoDB container to be created
	// Português: define o nome do container MongoDB a ser criado
	dockerContainer.SetContainerName("container_delete_mongo_after_test")

	// English: sets the value of the container's network port and the host port to be exposed
	// Português: define o valor da porta de rede do container e da porta do hospedeiro a ser exposta
	dockerContainer.AddPortToChange("27017", "27016")

	// English: sets MongoDB-specific environment variables (releases connections from any address)
	// Português: define variáveis de ambiente específicas do MongoDB (libera conexões de qualquer endereço)
	dockerContainer.SetEnvironmentVar(
		[]string{
			"--host 0.0.0.0",
		},
	)

	// English: sets the host computer's "./docker/mongodb/data/db" folder to the container's "/data/db" folder, so that the database data is archived on the host computer
	// Português: define a pasta "./docker/mongodb/data/db" do computador hospedeiro como sendo a pasta "/data/db" do container, para que os dados do banco de dados sejam arquivados no computador hospedeiro
	err = dockerContainer.AddFileOrFolderToLinkBetweenComputerHostAndContainer("./docker/mongodb/data/db", "/data/db")
	if err != nil {
		err = fmt.Errorf("dockerMongoDB.error: the function dockerContainer.AddFileOrFolderToLinkBetweenComputerHostAndContainer() returned an error: %v", err)
		return
	}

	// English: defines a text to be searched for in the standard output of the container indicating the end of the installation
	// define um texto a ser procurado na saída padrão do container indicando o fim da instalação
	dockerContainer.SetWaitStringWithTimeout(`"msg":"Waiting for connections","attr":{"port":27017`, 60*time.Second)

	// English: initialize the docker control object
	// Português: inicializa o objeto de controle docker
	err = dockerContainer.Init()
	if err != nil {
		err = fmt.Errorf("dockerMongoDB.error: the function dockerContainer.Init() returned an error: %v", err)
		return
	}

	// English: build a container
	// Português: monta o container
	err = dockerContainer.ContainerBuildAndStartFromImage()
	if err != nil {
		err = fmt.Errorf("dockerMongoDB.error: the function dockerContainer.ContainerBuildAndStartFromImage() returned an error: %v", err)
		return
	}

	return
}
```

## Download and parser pbf file

```golang
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
	var fileSaveName = "../commonFiles/sul-latest.osm.pbf"
	var fileTmpName = "../commonFiles/sul-latest.tmp"

	fmt.Println("Starting file download. This may take a while. It's ~300MB.")

	// English: Download the binary file with the map from Create Street Maps
	// Português: Faz o download do arquivo binário com o mapa do Create Street Maps
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
	err = compressData.Create(fileTmpName)
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
```
