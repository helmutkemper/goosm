package main

// Valid longitude values are between -180 and 180, both inclusive.
// Valid latitude  values are between  -90 and  90, both inclusive.

//https://rosettacode.org/wiki/Binary_search#Go

import (
	"context"
	"encoding/binary"
	"fmt"
	dockerBuilder "github.com/helmutkemper/iotmaker.docker.builder"
	dockerBuilderNetwork "github.com/helmutkemper/iotmaker.docker.builder.network"
	"github.com/pbnjay/memory"
	"github.com/qedus/osmpbf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	datasource "goosm/businessRules/dataSource"
	"goosm/compress"
	"goosm/goosm"
	downloadApiV06 "goosm/goosm/download"
	"io"
	"io/fs"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	KMainPodsCounter      = 1
	KSecondaryPodsCounter = 3
	KContainerTimeOut     = 2 * 60 * time.Second
	KMemorySwap           = 128 * dockerBuilder.KGigaByte
	KMemory               = 512 * dockerBuilder.KMegaByte
	// KLimit 131000 * osm.ConstructNode < 16MB of RAM (16MB is MongoDB document limit)
	KLimit = 131000
)

func main() {
	var err error

	// CPU Profile - code start
	//var profileFile *os.File
	//profileFile, err = os.Create("./cpuProfile.pb.gz")
	//if err != nil {
	//	panic(err)
	//}
	//err = pprof.StartCPUProfile(profileFile)
	//if err != nil {
	//	panic(err)
	//}
	//defer pprof.StopCPUProfile()
	// CPU Profile - code end

	//var pathToProcess = "./sul-latest.osm.pbf"

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//var memStats runtime.MemStats
	//runtime.ReadMemStats(&memStats)
	//
	//fmt.Printf("Memory in use: %d\n", memStats.Alloc)
	//fmt.Printf("Obtained from system: %d\n", memStats.Sys)

	log.Printf("Total system memory: %d MB", memory.TotalMemory())

	err = DockerSupport()
	if err != nil {
		panic(err)
	}

	err = datasource.Linker.Init(datasource.KMongoDB)
	if err != nil {
		panic(err)
	}

	//var way goosm.Way
	//way, err = datasource.Linker.Osm.GetWay(10492763, 5*time.Second)
	//if err != nil {
	//	return
	//}
	//p, _ := way.MakePolygonSurroundingsACW(50)
	//if p.PointInPolygon([2]float64{-48.51487113414021, -27.576612063557512}) {
	//	fmt.Printf("inside!\n")
	//}
	//fmt.Printf("%v", p.GeoJSonFeature)
	//
	////p, _ = way.MakePolygonSurroundingsACW(50)
	////if p.PointInPolygon([2]float64{-58.313311772251396, -33.091871001061975}) {
	////	fmt.Printf("inside!")
	////}
	//os.Exit(5)

	compressData := &compress.Compress{}
	compressData.Init(100)
	err = compressData.Open("./node.sul.tmp")
	if err != nil {
		panic(err)
	}
	defer compressData.Close()
	//_ = compressData.ReadFileHeaders()
	//_ = compressData.IndexToMemory()
	//err = compressData.ResizeBlock(1000)
	//if err != nil {
	//	panic(err)
	//}
	//os.Exit(7)

	total := uint64(0)
	start := time.Now()
	var binarySearch = &goosm.PbfProcess{}
	binarySearch.SetDatabase(datasource.Linker.Osm)
	binarySearch.SetDownloadApi(&downloadApiV06.DownloadApiV06{})
	binarySearch.SetDatabaseTimeout(10 * 60 * time.Second)
	binarySearch.SetCompress(compressData)
	// 119.32GB -
	//127.455.579.120
	//128.128.000.000B
	// 15h58m23.83964075s - total: 7.961.992.696
	// "./node.sul.tmp.bin" - id 10073799145
	//2022/11/02 13:00:55 main.go:129: tempo: 28m19.893329292s
	// 9m49.515609375s - total: 45248630
	total, err = binarySearch.CompleteParser("./sul-latest.osm.pbf")
	//total, err = binarySearch.SaveNodesIntoTmpFile("./sul-latest.osm.pbf")
	log.Printf("%v - total: %v", time.Since(start), total)
	if err != nil {
		panic(err)
	}

	//start = time.Now()
	//err = binarySearch.NodesToDatabase("./sul-latest.osm.pbf")
	//log.Printf("tempo: %v", time.Since(start))
	//if err != nil {
	//	panic(err)
	//}
	//
	//start = time.Now()
	err = binarySearch.WaysToDatabase("./sul-latest.osm.pbf")
	//log.Printf("tempo: %v", time.Since(start))
	//if err != nil {
	//	panic(err)
	//}
	os.Exit(0)

	query := bson.M{
		"$and": []bson.M{
			{
				"tag.boundary": bson.M{
					"$exists": true,
				},
			},
			{
				"admin_level": "3",
			},
			{
				"loc": bson.M{
					"$near": bson.M{
						"$geometry": bson.M{
							"type": "Point",
							"coordinates": []float64{
								-49.2167380,
								-25.4419983,
							},
							"maxDistance": 100 * 1000,
						},
					},
				},
			},
		},
	}
	_, geojson, _ := datasource.Linker.Osm.WayJoinQueryGeoJSonFeatures(query, 60*time.Second)
	log.Printf("%v", geojson)

	//way, _ := datasource.Linker.Osm.WayJoin(387270475, 5*time.Second)
	//_, geojson, _ = datasource.Linker.Osm.WayJoinGeoJSonFeatures(387270475, 5*time.Second)
	//log.Printf("%v", geojson)
	//
	//way.MakeGeoJSonFeature()
	//p, _ := way.MakePolygonSurroundings(50, 0)
	//
	//log.Printf("%v", p.GeoJSonFeature)
	//
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("total: %v", total)
	os.Exit(5)
	//err = binarySearch.WaysToDatabase(pathToProcess, "./node.tmp.bin")
	//log.Printf("%v", time.Since(start))
	//if err != nil {
	//	panic(err)
	//}

	// nodes: 45.248.630, ways: 4.108.131, relations: 78.943
	//var nodes, ways, relations uint64
	//nodes, ways, relations, err = OsmCounter(pathToProcess)
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("nodes: %v, ways: %v, relations: %v", nodes, ways, relations)

	//start := time.Now()
	//err = OsmWaySearchIds(pathToProcess, "./node.tmp.bin", int64(nodes))
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("%v", time.Since(start))

	//_, err = SaveWaysIntoFile(pathToProcess, "./way.tmp.bin", 1024*1024, 0)

	//err = SaveNodesIntoFile(pathToProcess, "./node.tmp.bin")
	//if err != nil {
	//	panic(err)
	//}

	//c := Compress{}
	//err = c.Compress(pathToProcess)
	//if err != nil {
	//	panic(err)
	//}

	//start := time.Now()
	//err = OsmNodeInstall(pathToProcess)
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("%v", time.Since(start))

	//start = time.Now()
	//err = OsmWayPreInstall(pathToProcess)
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("%v", time.Since(start))

	os.Exit(0)
	c := Compress{}

	//start = time.Now()
	err = c.loadNodesIntoFile("./node.tmp.bin")
	if err != nil {
		panic(err)
	}
	//log.Printf("%v", time.Since(start))
	//err = c.CompressV2(pathToProcess, "./tmp.bin")
	//if err != nil {
	//	panic(err)
	//}
}

func DockerSupport() (err error) {
	var netDocker *dockerBuilderNetwork.ContainerBuilderNetwork

	// Remove elementos docker residuais de testes anteriores
	// Rede, imagens, containers com o termo `delete` no nome
	// Perceba que isto não se aplica a imagem cache:latest
	err = dockerRemoveTestElements()
	if err != nil {
		return
	}

	// Cria uma rede docker
	// Como o gateway é 10.0.0.1, o primeiro endereço será 10.0.0.2
	netDocker, err = dockerTestNetworkCreate()
	if err != nil {
		return
	}

	var docker = new(dockerBuilder.ContainerBuilder)
	err = dockerMongoDB(netDocker, docker)
	return
}

func containerManager() (err error) {
	var netDocker *dockerBuilderNetwork.ContainerBuilderNetwork
	var containerMain = make([]*dockerBuilder.ContainerBuilder, KMainPodsCounter)
	var containerSecondary = make([]*dockerBuilder.ContainerBuilder, KSecondaryPodsCounter)

	var delta = 0

	// Remove elementos docker residuais de testes anteriores
	// Rede, imagens, containers com o termo `delete` no nome
	// Perceba que isto não se aplica a imagem cache:latest
	err = dockerRemoveTestElements()
	if err != nil {
		panic(err)
	}

	// Remove os elementos docker ao final do teste
	//defer func() {
	//	err = dockerRemoveTestElements()
	//	if err != nil {
	//		util.TraceToLog()
	//	}
	//}()

	// Cria, caso não exista, uma imagem de nome cache:latest
	// Veja TCModuleCache/simulation/cacheImage
	err = dockerBuildImageCache()
	if err != nil {
		panic(err)
	}

	// Cria uma rede docker
	// Como o gateway é 10.0.0.1, o primeiro endereço será 10.0.0.2
	netDocker, err = dockerTestNetworkCreate()
	if err != nil {
		panic(err)
	}

	err = dockerMakeImage(
		new(dockerBuilder.ContainerBuilder),
		"delete_map_image:latest",
		"./containerMaps",
	)
	if err != nil {
		panic(err)
	}

	err = dockerMakeImage(
		new(dockerBuilder.ContainerBuilder),
		"delete_secondary_pod_image:latest",
		"./test/simulation/syncDataBetweenPods_1/containerSecondary",
	)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i := int64(0); i != KMainPodsCounter; i += 1 {
		delta += 1
		log.Printf("delta: %v", delta)
		wg.Add(1)
		go func(i int64, wg *sync.WaitGroup) {
			var docker = &dockerBuilder.ContainerBuilder{}
			containerMain[i] = docker
			sufix := strconv.FormatInt(i, 10)
			err = dockerMakeContainer(
				netDocker,
				docker,
				"delete_main_pod_image:latest",
				"delete_main_pod_container"+sufix,
			)
			if err != nil {
				panic(err)
			}
			delta -= 1
			log.Printf("delta: %v", delta)
			wg.Done()
		}(i, &wg)
	}

	for i := int64(0); i != KSecondaryPodsCounter; i += 1 {
		delta += 1
		log.Printf("delta: %v", delta)
		wg.Add(1)
		go func(i int64, wg *sync.WaitGroup) {
			var docker = &dockerBuilder.ContainerBuilder{}
			containerSecondary[i] = docker
			sufix := strconv.FormatInt(i, 10)
			err = dockerMakeContainer(
				netDocker,
				docker,
				"delete_secondary_pod_image:latest",
				"delete_secondary_pod_container"+sufix,
			)
			if err != nil {
				panic(err)
			}
			delta -= 1
			log.Printf("delta: %v", delta)
			wg.Done()
		}(i, &wg)
	}

	log.Printf("wait delta: %v", delta)
	wg.Wait()

	for i := int64(0); i != KMainPodsCounter; i += 1 {
		delta += 1
		log.Printf("delta: %v", delta)
		wg.Add(1)
		go func(i int64, wg *sync.WaitGroup) {
			defer func(wg *sync.WaitGroup) {
				delta -= 1
				log.Printf("delta: %v", delta)
				wg.Done()
			}(wg)

			err = containerMain[i].ContainerStartAfterBuild()
			if err != nil {
				panic(err)
			}

			_, err = containerMain[i].WaitForTextInContainerLogWithTimeout("main container finished", KContainerTimeOut)
			if err != nil {
				panic(err)
			}
			log.Printf("done: main container finished")

		}(i, &wg)
	}

	for i := int64(0); i != KSecondaryPodsCounter; i += 1 {
		delta += 1
		log.Printf("delta: %v", delta)
		wg.Add(1)
		go func(i int64, wg *sync.WaitGroup) {
			defer func(wg *sync.WaitGroup) {
				delta -= 1
				log.Printf("delta: %v", delta)
				wg.Done()
			}(wg)

			err = containerSecondary[i].ContainerStartAfterBuild()
			if err != nil {
				panic(err)
			}

			_, err = containerSecondary[i].WaitForTextInContainerLogWithTimeout("secondary container started", KContainerTimeOut)
			if err != nil {
				panic(err)
			}

			log.Printf("Esperando para pausar container")
			time.Sleep(2 * time.Second)

			for p := 0; p != 15; p += 1 {
				log.Printf("Container pausado")
				err = containerSecondary[i].ContainerPause()
				if err != nil {
					panic(err)
				}

				log.Printf("Esperando para continuar o container")
				time.Sleep(2 * time.Second)

				log.Printf("Container continuando")
				err = containerSecondary[i].ContainerUnpause()
				if err != nil {
					panic(err)
				}
				time.Sleep(2 * time.Second)
			}

			log.Printf("Esperando o sinal de encerramento do container")
			_, err = containerSecondary[i].WaitForTextInContainerLogWithTimeout("secondary container finished", KContainerTimeOut)
			if err != nil {
				panic(err)
			}

			log.Printf("done: secondary container finished")

		}(i, &wg)
	}

	log.Printf("wait delta: %v", delta)
	wg.Wait()

	var containsText bool
	for _, container := range containerMain {
		containsText, err = container.FindTextInsideContainerLog("Total: 50000")
		if err != nil {
			panic(err)
		}
		if containsText != true {
			panic(err)
		}
	}

	for _, container := range containerSecondary {
		containsText, err = container.FindTextInsideContainerLog("Total: 50000")
		if err != nil {
			panic(err)
		}
		if containsText != true {
			panic(err)
		}
	}

	return
}

// dockerTestNetworkCreate (português): Cria uma rede docker para as simulações.
//
//	Saída:
//	  netDocker: Ponteiro para o objeto gerenciador de rede docker
//	  err: Objeto padrão de erro
func dockerTestNetworkCreate() (
	netDocker *dockerBuilderNetwork.ContainerBuilderNetwork,
	err error,
) {

	log.Print("instalação da rede de teste no docker: início")

	// Cria um orquestrador de rede para o container [opcional]
	netDocker = &dockerBuilderNetwork.ContainerBuilderNetwork{}
	err = netDocker.Init()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	// Cria uma rede de nome "cache_delete_after_test"
	err = netDocker.NetworkCreate(
		"cache_delete_after_test",
		"10.0.0.0/16",
		"10.0.0.1",
	)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	log.Print("instalação da rede de teste no docker: fim")

	return
}

// dockerBuildImageCache (português): Cria, caso não exista, uma imagem de nome
// `cache:latest` para ser usada pelo código.
//
//	Saída:
//	  err: Objeto padrão de erro
func dockerBuildImageCache() (err error) {

	log.Print("verificando a necessidade de criação da imagem cache:latest: início")
	defer log.Print("verificando a necessidade de criação da imagem cache:latest: fim")

	// Este é o nome padrão da imagem cache usada pelo projeto docker.build
	// Embora esta cache seja grande, ela é usada apenas na primeira etapa do build e não
	// interface no tamanho total do container final. Ela apenas contém os elementos para
	// o processo de build.
	var imageCacheName = "cache:latest"
	var imageId string
	var container = &dockerBuilder.ContainerBuilder{}

	// Caso a imagem exista, ignora o resto do código
	imageId, err = container.ImageFindIdByName(imageCacheName)
	if err != nil && err.Error() != "image name not found" {
		return
	}

	if imageId != "" {
		return
	}

	// Define o nome da imagem
	container.SetImageName(imageCacheName)
	// Imprime a saída padrão do container na saída padrão do golang
	container.SetPrintBuildOnStrOut()
	// fixme: desnecessário?
	//container.SetContainerName(imageCacheName)
	// Caminho relativo da pasta usada na criação da imagem
	container.SetBuildFolderPath("./cacheImage")
	// Inicializa o objeto dockerBuilder.ContainerBuilder
	err = container.Init()
	if err != nil {
		return
	}
	// Monta a imagem baseada no conteúdo da pasta
	_, err = container.ImageBuildFromFolder()
	if err != nil {
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

	// set a docker network
	dockerContainer.SetNetworkDocker(netDocker)
	// set an image name
	dockerContainer.SetImageName("mongo:latest")
	// set a container name
	dockerContainer.SetContainerName("container_delete_mongo_after_test")
	// set a port to expose
	//dockerContainer.AddPortToExpose("27017")
	dockerContainer.AddPortToChange("27017", "27016")
	// se a environment var list
	dockerContainer.SetEnvironmentVar(
		[]string{
			"--host 0.0.0.0",
		},
	)
	// set a MongoDB data dir to ./test/data
	err = dockerContainer.AddFileOrFolderToLinkBetweenComputerHostAndContainer("./docker/mongodb/data/db", "/data/db")
	if err != nil {
		return
	}
	// set a text indicating for container ready for use
	dockerContainer.SetWaitStringWithTimeout(`"msg":"Waiting for connections","attr":{"port":27017`, 20*time.Second)

	// inicialize the object before sets
	err = dockerContainer.Init()
	if err != nil {
		return
	}

	// build a container
	err = dockerContainer.ContainerBuildAndStartFromImage()
	return
}

func MongoDbMakeCollection(
	address string,
	timeout time.Duration,
) (
	collection *mongo.Collection,
	err error,
) {
	var mongoClient *mongo.Client
	var cancel context.CancelFunc
	var ctx context.Context

	// (English): Prepare the MongoDB client
	// (Português): Prepara o cliente do MongoDB
	mongoClient, err = mongo.NewClient(options.Client().ApplyURI(address))
	if err != nil {
		return
	}

	// (English): Connects to MongoDB
	// (Português): Conecta ao MongoDB
	err = mongoClient.Connect(ctx)
	if err != nil {
		return
	}

	// (English): Prepares the timeout
	// (Português): Prepara o tempo limite
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// (English): Ping() to test the MongoDB connection
	// (Português): Faz um ping() para testar a conexão do MongoDB
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return
	}

	// (English): Creates the 'test' bank and the 'dinos' collection
	// (Português): Cria o banco 'test' e a coleção 'dinos'
	collection = mongoClient.Database("osm").Collection("original")
	return
}

func dockerMakeImage(
	dockerContainer *dockerBuilder.ContainerBuilder,
	imageName string,
	folderPath string,
) (
	err error,
) {

	log.Printf("image %v build: start", imageName)
	defer log.Printf("image %v build: end", imageName)

	// Imprime a saída padrão do container na saída padrão do golang
	dockerContainer.SetPrintBuildOnStrOut()
	// Habilita o uso da imagem cache:latest
	dockerContainer.SetCacheEnable(true)
	// Determina o nome da imagem a ser usada
	dockerContainer.SetImageName(imageName)
	// Define o caminho da pasta contando o projeto
	dockerContainer.SetBuildFolderPath(folderPath)
	// Gera o Dockerfile de forma automática
	dockerContainer.MakeDefaultDockerfileForMe()
	// Define os repositórios da TC como sendo privados
	dockerContainer.SetGitPathPrivateRepository("github.com/tradersclub")

	dockerContainer.SetImageBuildOptionsCPUPeriod(100000)
	dockerContainer.SetImageBuildOptionsCPUQuota(100000)
	//dockerContainer.SetImageBuildOptionsCPUSetCPUs("1-8")
	//dockerContainer.SetImageBuildOptionsMemorySwap(KMemorySwap)
	dockerContainer.SetImageBuildOptionsMemory(KMemory)

	// Copias as credenciais do usuário para o container
	err = dockerContainer.SetPrivateRepositoryAutoConfig()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	// Inicializa o objeto dockerBuilder.ContainerBuilder
	err = dockerContainer.Init()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}
	// Gera a imagem baseada no conteúdo da pasta
	_, err = dockerContainer.ImageBuildFromFolder()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	return
}

func dockerMakeContainer(
	netDocker *dockerBuilderNetwork.ContainerBuilderNetwork,
	dockerContainer *dockerBuilder.ContainerBuilder,
	imageName string,
	containerName string,
) (
	err error,
) {

	log.Printf("container %v build: start", containerName)
	defer log.Printf("container %v build: end", containerName)

	// Imprime a saída padrão do container na saída padrão do golang
	dockerContainer.SetPrintBuildOnStrOut()
	// Habilita o uso da imagem cache:latest
	dockerContainer.SetCacheEnable(true)
	// Aponta o gerenciador de rede
	dockerContainer.SetNetworkDocker(netDocker)
	// Determina o nome da imagem a ser usada
	dockerContainer.SetImageName(imageName)
	// Define o nome do container
	dockerContainer.SetContainerName(containerName)
	// Gera o Dockerfile de forma automática
	dockerContainer.MakeDefaultDockerfileForMe()
	// Define os repositórios da TC como sendo privados
	dockerContainer.SetGitPathPrivateRepository("github.com/tradersclub")

	dockerContainer.SetImageBuildOptionsCPUPeriod(100000)
	dockerContainer.SetImageBuildOptionsCPUQuota(100000)
	//dockerContainer.SetImageBuildOptionsCPUSetCPUs("1-8")
	//dockerContainer.SetImageBuildOptionsMemorySwap(KMemorySwap)
	dockerContainer.SetImageBuildOptionsMemory(KMemory)

	// Copias as credenciais do usuário para o container
	err = dockerContainer.SetPrivateRepositoryAutoConfig()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	// Inicializa o objeto dockerBuilder.ContainerBuilder
	err = dockerContainer.Init()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}
	// Gera o container
	err = dockerContainer.ContainerBuildWithoutStartingItFromImage()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	return
}

// dockerRemoveTestElements (português): Remove qualquer elemento docker com o termo
// `delete` no nome
//
//	Saída:
//	  err: Objeto de erro padrão
func dockerRemoveTestElements() (err error) {

	log.Print("removendo rede, container e imagens docker com o termo delete no nome: início")

	// Cria um objeto coletor de lixo
	var garbageCollector = dockerBuilder.ContainerBuilder{}
	err = garbageCollector.Init()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	// Procura por redes, containers, volumes e imagens com o termo "delete" no nome e apaga
	err = garbageCollector.RemoveAllByNameContains("delete")
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	log.Print("removendo rede, container e imagens docker com o termo delete no nome: fim")
	return
}

// MakeCacheImage
//
// Cria, caso não exista, uma imagem de nome cache:latest
//
//	Saída:
//	  err: objeto padrão de erro do golang
//
// Nota: - cache:latest é o nome padrão da imagem cache usada pelo projeto docker.build
//   - Embora esta cache seja grande, ela é usada apenas na primeira etapa do build e não interfere no tamanho total
//     do container final. Ela apenas contém os elementos para o processo de build.
func MakeCacheImage() (err error) {

	var imageCache = &dockerBuilder.ContainerBuilder{}
	// Imprime a saída padrão do container na saída padrão do golang ajudando a seguir o processode build.
	imageCache.SetPrintBuildOnStrOut()
	// Define o nome da imagem
	imageCache.SetImageName("cache:latest")
	// Caminho relativo da pasta contendo o Dockerfile da imagem a ser criada
	imageCache.SetBuildFolderPath("../../../test/cacheImage")
	// Determina a validade da imagem.
	// Se o intervalo entre a data atual e a data de criação da imagem for menos do que o valor indicado, não será criada
	// uma nova imagem.
	imageCache.SetImageExpirationTime(365 * 24 * time.Hour)

	// Inicializa o objeto docker builder
	err = imageCache.Init()
	if err != nil {
		log.Printf("imageCache.Init().error: %v", err)
		return
	}

	// Monta a imagem a partir do conteúdo da pasta indicada acima
	_, err = imageCache.ImageBuildFromFolder()
	if err != nil {
		log.Printf("imageCache.ImageBuildFromFolder().error: %v", err)
		return
	}

	return
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func GetMapAddress(id int64) (address int64) {
	return id / KLimit
}

// OsmNodeInstall
//
// English:
//
// Installs all points of interest in the database.
//
//	Notes:
//	  * Points of interest not all node types with non-empty tag property.
//
// Português:
//
// Instala todos os pontos de interesse no banco de dados.
//
//	Notas:
//	  * Pontos de interesse não todos os tipos nodes com a propriedade tag não vazia.
func OsmNodeInstall(osmFilePath string) (err error) {

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		return
	}

	defer func() {
		err := osmFile.Close()
		if err != nil {
			log.Printf("error closing main osm source file: %v", err.Error())
		}
	}()

	osmDecoder := osmpbf.NewDecoder(osmFile)

	// use more memory from the start, it is faster
	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		return
	}

	var nodeList = make([]goosm.Node, 0)

	var nc, wc, rc uint64
	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:
				nc++

				if converted.Info.Visible && len(converted.Tags) != 0 {

					if len(converted.Tags) != 0 {
						node := goosm.Node{}
						node.Init(converted.ID, converted.Lon, converted.Lat, &converted.Tags)
						if len(node.Tag) == 0 {
							continue
						}

						nodeList = append(nodeList, node)
						if len(nodeList) == 500 {
							err = datasource.Linker.Osm.SetNode(&nodeList)
							if err != nil {
								return
							}

							nodeList = make([]goosm.Node, 0)
						}
					}

				}

			case *osmpbf.Way:
				wc++

				if nodeList != nil && len(nodeList) != 0 {
					err = datasource.Linker.Osm.SetNode(&nodeList)
					if err != nil {
						return
					}

					nodeList = nil
				}

				return
			}
		}
	}

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
	return
}

func removeByIndex(data *[]int64, index int) {
	if index == -1 {
		return
	}

	*data = append((*data)[:index], (*data)[index+1:]...)
}

func findIndex(data *[]int64, value int64) (index int) {
	for k, v := range *data {
		if v == value {
			return k
		}
	}

	return -1
}

// OsmWayPreInstall
//
// English:
//
// Installs all found ways that are visible, but does not process them.
//
// Português:
//
// Instala todos os ways encontrados, que sejam visíveis, porém, não os processa.
//func OsmWayPreInstall(osmFilePath string) (err error) {
//
//	var osmFile *os.File
//	osmFile, err = os.Open(osmFilePath)
//	if err != nil {
//		return
//	}
//
//	defer func() {
//		err := osmFile.Close()
//		if err != nil {
//			log.Printf("error closing main osm source file: %v", err.Error())
//		}
//	}()
//
//	osmDecoder := osmpbf.NewDecoder(osmFile)
//
//	// use more memory from the start, it is faster
//	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)
//
//	// start decoding with several goroutines, it is faster
//	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
//	if err != nil {
//		return
//	}
//
//	var wayList = make([]goosm.Way, 0)
//
//	for {
//		var osmPbfElement interface{}
//		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
//			err = nil
//			break
//		} else if err != nil {
//			return
//		} else {
//			switch converted := osmPbfElement.(type) {
//
//			case *osmpbf.Way:
//				if converted.Info.Visible == false {
//					continue
//				}
//
//				way := goosm.Way{}
//				way.Id = converted.ID
//				way.Rad = make([][2]float64, len(converted.NodeIDs))
//				way.Loc = make([][2]float64, len(converted.NodeIDs))
//				way.IdList = converted.NodeIDs
//				way.Tag = converted.Tags
//				//err = way.Init()
//				//if err != nil {
//				//	return
//				//}
//
//				wayList = append(wayList, way)
//				if len(wayList) == 500 {
//					err = datasource.Linker.Osm.SetWay(&wayList, 5*time.Second)
//					if err != nil {
//						return
//					}
//
//					wayList = make([]goosm.Way, 0)
//				}
//
//			case *osmpbf.Relation:
//
//				if wayList != nil && len(wayList) != 0 {
//					err = datasource.Linker.Osm.SetWay(&wayList, 5*time.Second)
//					if err != nil {
//						return
//					}
//
//					wayList = nil
//				}
//				return
//			}
//		}
//	}
//
//	return
//}

//func OsmInstall(osmFilePath string) (err error) {
//
//	//start := time.Now()
//
//	var osmFile *os.File
//	osmFile, err = os.Open(osmFilePath)
//	if err != nil {
//		return
//	}
//
//	defer func() {
//		err := osmFile.Close()
//		if err != nil {
//			log.Printf("error closing main osm source file: %v", err.Error())
//		}
//	}()
//
//	osmDecoder := osmpbf.NewDecoder(osmFile)
//
//	// use more memory from the start, it is faster
//	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)
//
//	// start decoding with several goroutines, it is faster
//	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
//	if err != nil {
//		return
//	}
//
//	//var data osm.DataMap
//
//	var nodeList = make([]goosm.Node, 0)
//	var wayList = make([]goosm.Way, 0)
//
//	var nc, wc, rc uint64
//	for {
//		var osmPbfElement interface{}
//		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
//			err = nil
//			break
//		} else if err != nil {
//			return
//		} else {
//			switch converted := osmPbfElement.(type) {
//			case *osmpbf.Node:
//				nc++
//
//				if converted.Info.Visible && len(converted.Tags) != 0 {
//
//					if len(converted.Tags) != 0 {
//						node := goosm.Node{}
//						node.Init(converted.ID, converted.Lon, converted.Lat, &converted.Tags)
//						if len(node.Tag) == 0 {
//							continue
//						}
//
//						nodeList = append(nodeList, node)
//						if len(nodeList) == 500 {
//							err = datasource.Linker.Osm.SetNode(&nodeList)
//							if err != nil {
//								return
//							}
//
//							nodeList = make([]goosm.Node, 0)
//						}
//					}
//
//				}
//
//			case *osmpbf.Way:
//				wc++
//
//				if nodeList != nil && len(nodeList) != 0 {
//					err = datasource.Linker.Osm.SetNode(&nodeList)
//					if err != nil {
//						return
//					}
//
//					nodeList = nil
//				}
//
//				if converted.Info.Visible == false {
//					continue
//				}
//
//				way := goosm.Way{}
//				way.Id = converted.ID
//				way.Loc = make([][2]float64, 0)
//				way.Rad = make([][2]float64, 0)
//				way.IdList = converted.NodeIDs
//				way.Tag = converted.Tags
//				//err = way.Init()
//				//if err != nil {
//				//	return
//				//}
//
//				wayList = append(wayList, way)
//				if len(wayList) == 500 {
//					start := time.Now()
//					err = datasource.Linker.Osm.SetWay(&wayList, 5*time.Second)
//					if err != nil {
//						return
//					}
//					log.Printf("%v", time.Since(start))
//
//					wayList = make([]goosm.Way, 0)
//				}
//
//			case *osmpbf.Relation:
//
//				// Process Relation v.
//				rc++
//
//				if wayList != nil && len(wayList) != 0 {
//					err = datasource.Linker.Osm.SetWay(&wayList, 5*time.Second)
//					if err != nil {
//						return
//					}
//
//					wayList = nil
//				}
//				return
//
//			default:
//				log.Fatalf("unknown type %T\n", converted)
//			}
//		}
//	}
//
//	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
//	return
//}

type Compress struct {
	id         []byte `bson:"-"`
	longitude  []byte `bson:"-"`
	latitude   []byte `bson:"-"`
	memory     map[uint64][2]float64
	memoryPage int64
}

//func (e *Compress) Compress(osmFilePath string) (err error) {
//
//	var osmFile *os.File
//	osmFile, err = os.Open(osmFilePath)
//	if err != nil {
//		return
//	}
//
//	defer func() {
//		err := osmFile.Close()
//		if err != nil {
//			log.Printf("error closing main osm source file: %v", err.Error())
//		}
//	}()
//
//	osmDecoder := osmpbf.NewDecoder(osmFile)
//
//	// use more memory from the start, it is faster
//	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)
//
//	// start decoding with several goroutines, it is faster
//	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
//	if err != nil {
//		return
//	}
//
//	e.id = make([]byte, 8)
//	e.latitude = make([]byte, 8)
//	e.longitude = make([]byte, 8)
//	counter := 0
//
//	// 512 * 1024 is based on 16MB limit from MongoDB document size
//	max := 1024
//	compressed := osm.Compressed{}
//	compressed.IdList = make([]int64, 0)
//	compressed.Data = make(map[int64][2]float64)
//
//	for {
//		var osmPbfElement interface{}
//		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
//			err = nil
//			break
//		} else if err != nil {
//			return
//		} else {
//			switch converted := osmPbfElement.(type) {
//			case *osmpbf.Node:
//
//				if counter >= max {
//
//					compressed.Id = primitive.NewObjectID()
//					err = datasource.Linker.Osm.SetCompress(&compressed)
//					if err != nil {
//						return
//					}
//
//					counter = 0
//					//compressed.Data = make([]byte, max)
//					compressed.IdList = make([]int64, 0)
//					compressed.Data = make(map[int64][2]float64)
//				}
//
//				//binary.LittleEndian.PutUint64(e.id, uint64(converted.ID))
//				//vLat := int64(converted.Lat * 10000000)
//				//binary.LittleEndian.PutUint64(e.latitude, uint64(vLat))
//				//vLon := int64(converted.Lon * 10000000)
//				//binary.LittleEndian.PutUint64(e.longitude, uint64(vLon))
//				//
//				//compressed.Data = append(
//				//	compressed.Data,
//				//	e.id[0], e.id[1], e.id[2], e.id[3], e.id[4], e.id[5], e.id[6], e.id[7],
//				//	e.latitude[0], e.latitude[1], e.latitude[2], e.latitude[3], e.latitude[4], e.latitude[5], e.latitude[6], e.latitude[7],
//				//	e.longitude[0], e.longitude[1], e.longitude[2], e.longitude[3], e.longitude[4], e.longitude[5], e.longitude[6], e.longitude[7],
//				//)
//
//				compressed.IdList = append(compressed.IdList, converted.ID)
//				compressed.Data[converted.ID] = [2]float64{converted.Lon, converted.Lat}
//
//				counter++
//
//			default:
//
//				compressed.Id = primitive.NewObjectID()
//				err = datasource.Linker.Osm.SetCompress(&compressed)
//				if err != nil {
//					return
//				}
//
//				return
//			}
//		}
//	}
//
//	return
//}

func (e *Compress) loadNodesIntoFile(tmpFilePath string) (err error) {
	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_RDONLY, fs.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	var byteSize = int64(8)
	var filePointer = int64(0)
	var idAsByte = make([]byte, 8)
	var lonAsByte = make([]byte, 8)
	var latAsByte = make([]byte, 8)
	var idAsInt int64
	var loc = [2]float64{}

	for {
		_, err = tmpFile.ReadAt(idAsByte, filePointer)
		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		filePointer += byteSize
		v := binary.LittleEndian.Uint64(idAsByte)
		idAsInt = int64(v)

		_, err = tmpFile.ReadAt(lonAsByte, filePointer)
		filePointer += byteSize
		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		filePointer += byteSize
		v = binary.LittleEndian.Uint64(lonAsByte)
		loc[0] = float64(int64(v)) / 10000000.0

		_, err = tmpFile.ReadAt(latAsByte, filePointer)
		filePointer += byteSize
		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		filePointer += byteSize
		v = binary.LittleEndian.Uint64(latAsByte)
		loc[1] = float64(int64(v)) / 10000000.0
	}

	_ = idAsInt
	return
}

// desnecessário??????
func SaveWaysIntoFile(osmFilePath, tmpFilePath string, limit, page int64) (end bool, err error) {

	_ = os.Remove(tmpFilePath)

	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY, fs.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		return
	}

	defer func() {
		err := osmFile.Close()
		if err != nil {
			log.Printf("error closing main osm source file: %v", err.Error())
		}
	}()

	osmDecoder := osmpbf.NewDecoder(osmFile)

	// use more memory from the start, it is faster
	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		return
	}

	var data = make([]byte, 8)
	skip := limit * page // 2 * 1
	counter := int64(0)

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:
				continue

			case *osmpbf.Way:

				if counter < skip {
					counter++
					continue
				}

				if counter >= skip+limit {
					return
				}

				counter++

				for _, wayID := range converted.NodeIDs {
					v := uint64(wayID)
					binary.LittleEndian.PutUint64(data, v)

					// Id do way contido no node
					_, err = tmpFile.Write(data)
					if err != nil {
						return
					}

					// Espaço para longitude
					_, err = tmpFile.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
					if err != nil {
						return
					}

					// Espaço para latitude
					_, err = tmpFile.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
					if err != nil {
						return
					}
				}

			default:
				end = true
				break
			}
		}
	}

	return
}

func (e *Compress) saveWaysIdIntoFile(osmFilePath, tmpFilePath string) (err error) {

	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY, fs.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		return
	}

	defer func() {
		err := osmFile.Close()
		if err != nil {
			log.Printf("error closing main osm source file: %v", err.Error())
		}
	}()

	osmDecoder := osmpbf.NewDecoder(osmFile)

	// use more memory from the start, it is faster
	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		return
	}

	var id = make([]byte, 8)
	var zero = make([]byte, 8)

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:
				continue

			case *osmpbf.Way:

				for _, wayId := range converted.NodeIDs {
					binary.LittleEndian.PutUint64(id, uint64(wayId))

					// Id do way contido no node
					_, err = tmpFile.Write(id)
					if err != nil {
						return
					}

					// Espaço para longitude
					_, err = tmpFile.Write(zero)
					if err != nil {
						return
					}

					// Espaço para latitude
					_, err = tmpFile.Write(zero)
					if err != nil {
						return
					}
				}

			default:
				break
			}
		}
	}

	return
}

func (e *Compress) saveNodesIntoMemory(osmFilePath string, page int64) (err error) {

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		return
	}

	defer func() {
		err := osmFile.Close()
		if err != nil {
			log.Printf("error closing main osm source file: %v", err.Error())
		}
	}()

	osmDecoder := osmpbf.NewDecoder(osmFile)

	// use more memory from the start, it is faster
	osmDecoder.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = osmDecoder.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		return
	}

	limit := 8 * 1024 * 1024
	currentPage := int64(0)
	counter := 0
	e.memory = make(map[uint64][2]float64)
	pass := false

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:

				if currentPage == page {
					pass = true
					e.memory[uint64(converted.ID)] = [2]float64{converted.Lon, converted.Lat}
				}

				if currentPage > page {
					return
				}

				counter++
				if counter > limit {
					currentPage++
				}

				continue

			default:

				if pass == false {
					err = io.EOF
					return
				}

				break
			}
		}
	}

	return
}

func (e *Compress) saveNodesIdIntoFile(tmpFilePath string) (err error) {

	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_RDWR, fs.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	var idAsByte = make([]byte, 8)
	var LonAsByte = make([]byte, 8)
	var LatAsByte = make([]byte, 8)
	var idAsUInt64 uint64
	var found bool
	var location [2]float64

	filePointer := int64(0)

	for {
		_, err = tmpFile.ReadAt(idAsByte, filePointer)
		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		idAsUInt64 = binary.LittleEndian.Uint64(idAsByte)
		if location, found = e.memory[idAsUInt64]; found {
			vLon := int64(location[0] * 10000000)
			binary.LittleEndian.PutUint64(LonAsByte, uint64(vLon))
			vLat := int64(location[1] * 10000000)
			binary.LittleEndian.PutUint64(LatAsByte, uint64(vLat))

			filePointer += 8
			_, err = tmpFile.WriteAt(LonAsByte, filePointer)
			if err != nil {
				return
			}

			filePointer += 8
			_, err = tmpFile.WriteAt(LatAsByte, filePointer)
			if err != nil {
				return
			}

			// next id
			filePointer += 8
		} else {
			filePointer += 8 // +8 longitude
			filePointer += 8 // +8 latitude
			filePointer += 8 // +8 next id
		}
	}
}

func (e *Compress) CompressV2(osmFilePath string, tmpFilePath string) (err error) {
	err = e.saveWaysIdIntoFile(osmFilePath, tmpFilePath)
	if err != nil {
		return
	}

	page := int64(0)
	for {
		err = e.saveNodesIntoMemory(osmFilePath, page)
		if err != nil {
			return
		}

		err = e.saveNodesIdIntoFile(tmpFilePath)
		if err != nil {
			return
		}

		page++
	}
}
