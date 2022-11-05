package goosm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/qedus/osmpbf"
	"goosm/module/util"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"time"
)

type CompressInterface interface {
	// Init
	//
	// English:
	//
	// Initializes the object.
	//
	//   Input:
	//     blockSize: Spacing between ID captures for the in-memory index.
	//
	//   Note:
	//     * Very small values consume a lot of memory and make the file very large, but generate greater efficiency.
	//     * For the planet file, with ~8 billion of points, 8000000000 nodes / 1000 block size * 8 bytes per ID = 61MB
	//
	// Português:
	//
	// Inicializa o objeto.
	//
	//   Entrada:
	//     blockSize: Espaçamento entre as capturas de IDs para o índice em memória.
	//
	//   Nota:
	//     * Valores muitos pequenos consomem muita memória e deixam o arquivo muito grande, mas, geram maior eficiência.
	//     * Para o arquivo do planeta, com ~8 bilhões de pontos, 8000000000 nodes / 1000 block size * 8 bytes por ID = 61MB
	Init(blockSize int64)

	// Round
	//
	// English:
	//
	// Rounds a floating point to N decimal places
	//
	//  Input:
	//    value: value to be rounded off;
	//    places: number of decimal places. Eg. 7.0
	//
	// Português:
	//
	// Rounds a floating point to N decimal places
	//
	//  Entrada:
	//    value: valor a ser arredondado;
	//    places: quantidade de casas decimais. Ex: 7.0
	Round(value, places float64) float64

	// Open
	//
	// English:
	//
	// Open the temporary file.
	//
	// Português:
	//
	// Abre o arquivo temporário.
	Open(path string) (err error)

	// Close
	//
	// English:
	//
	// Close the temporary file
	//
	// Português:
	//
	// Fecha o arquivo temporário
	Close()

	// WriteNode
	//
	// English:
	//
	// Write node to temporary file.
	//
	//  Input:
	//    id: positive number greater than zero;
	//    longitude: value between ±180 to 7 decimal places;
	//    latitude: value between ±90 with 7 decimal places;
	//
	//  Notas:
	//    * Node format saved 8 bytes for ID, 4 bytes for longitude and 4 bytes for latitude;
	//    * Longitude and latitude save 7 decimal places;
	//    * The most significant bit is 1 indicates a negative number, but the two's complement rule was not used.
	//      If the most significant bit is 1, the number is negative, otherwise positive.
	//    * Coordinate compression at 4 bytes saves considerable time and compresses the final file.
	//
	// Português:
	//
	// Escreve o node no arquivo temporário.
	//
	//  Entrada:
	//    id: número positivo maior do que zero;
	//    longitude: valor entre ±180 com 7 casas decimais;
	//    latitude: valor entre ±90 com 7 casas decimais;
	//
	//  Notes:
	//    * Formato do node salvo 8 bytes para ID, 4 bytes para longitude e 4 bytes para latitude;
	//    * Longitude e latitude salvam 7 casas decimais;
	//    * O bit mais significativo em 1, indica um número negativo, mas, não foi usada a regra do complemento de dois.
	//      Caso o bit mais significativo seja 1, o número é negativo, caso contrário, positivo.
	//    * A compactação de coordenada em 4 bytes gera um ganho de tempo considerável e compacta o arquivo final.
	WriteNode(id int64, longitude, latitude float64) (err error)

	// FindNodeByID
	//
	// English:
	//
	// Search for longitude and latitude in the temporary file.
	//
	//  Input:
	//    id: ID of the node sought.
	//
	//  Output:
	//    longitude: value between ±180 width 7 decimal places;
	//    latitude: value between ±90 width 7 decimal places;
	//    err: pattern object, with io.EOF error when value not found in file
	//
	// Português:
	//
	// Procura por longitude e latitude no arquivo temporário.
	//
	//  Entrada:
	//    id: ID do node procurado.
	//
	//  Saída:
	//    longitude: valor entre ±180 com 7 casas decimais;
	//    latitude: valor entre ±90 com 7 casas decimais;
	//    err: objeto de padrão, com erro io.EOF quando o valor não é encontrado no arquivo
	FindNodeByID(id int64) (longitude, latitude float64, err error)

	// IndexToMemory
	//
	// English:
	//
	// Loads the indexes contained in the temporary file into memory.
	//
	// Português:
	//
	// Carrega os índices contidos no arquivo temporário na memória.
	IndexToMemory() (err error)

	// MountIndexIntoFile
	//
	// English:
	//
	// Salva os índices no arquivo temporário.
	// See the explanation on the Init() function for more details.
	//
	//  Note:
	//    * Indexes are blocks with ranges of IDs to help calculate the address of the ID within the temporary file.
	//    * Indexes are loaded into memory for better performance.
	//
	// Português:
	//
	// Salva os índices no arquivo temporário.
	// Veja a explicação na função Init() para mais detalhes.
	//
	//  Nota:
	//    * Índices são blocos com intervalos de IDs para ajudar a calcular o endereço do ID dentro do arquivo temporário.
	//    * Índices são carregados em memória para maior desempenho.
	MountIndexIntoFile() (err error)

	// WriteFileHeaders
	//
	// English:
	//
	// Write configuration data at the beginning of the file.
	//
	// Português:
	//
	// Escreve os dados de configuração no início do arquivo.
	WriteFileHeaders() (err error)

	// ReadFileHeaders
	//
	// English:
	//
	// Read the configuration data at the beginning of the file.
	//
	// Português:
	//
	// Lê os dados de configuração no início do arquivo.
	ReadFileHeaders() (err error)
}

type InterfaceDownloadOsm interface {
	DownloadNode(id int64) (node Node, err error)
	DownloadWay(id int64) (way Way, err error)
	DownloadRelation(id int64) (relation Relation, err error)
}

type InterfaceDatabase interface {
	SetTimeout(timeout time.Duration)
	Connect(connectionString string, args ...interface{}) (err error)
	Close() (err error)
	New() (referenceInitialized interface{}, err error)
	SetNodeOne(node *Node) (err error)
	SetNodeMany(list *[]Node) (err error)
	SetWrongWayNodeMany(list *[]Node) (err error)
	GetNodeById(id int64) (node Node, err error)

	SetWay(wayList *[]Way, timeout time.Duration) (err error)
	//GetWay(id int64, timeout time.Duration) (way Way, err error)
	GetWayById(id int64) (way Way, err error)
	//WayJoin(id int64, timeout time.Duration) (way Way, err error)

	// WayJoinGeoJSonFeatures
	//
	// English:
	//
	// Joins all the forming segments of a way that has the "tag.name" property defined and returns the geoJSon of the same
	//
	//  Input:
	//    id: ID of any segment forming the set;
	//    timeout: timeout of each database request.
	//
	//  Output:
	//    distanceMeters: total distance in meters (forks are counted);
	//    features: list of geojson features;
	//    err: golang's default error object.
	//
	// Português:
	//
	// Junta todos os seguimentos formadores de um way que tenha a propriedade "tag.name" definida e retorna o geoJSon do
	// mesmo
	//
	//  Entrada:
	//    id: ID de qualquer seguimento formador do conjunto;
	//    timeout: tempo limit de cada requisição do banco de dados.
	//
	//  Saída:
	//    distanceMeters: distância total em metros (bifurcações são contadas);
	//    features: lista de geojson features;
	//    err: objeto padrão de erro do golang.
	WayJoinGeoJSonFeatures(id int64, timeout time.Duration) (distanceMeters float64, features string, err error)

	WayJoinQueryGeoJSonFeatures(query interface{}, timeout time.Duration) (distanceMeters float64, features string, err error)
}

// PbfProcess
//
// English:
//
// Binary search is used to populate the database because of the database response time curve.
//
// The problem: as the database is being populated with construction-only nodes, the insertion/search time in the
// database goes up a lot and Open Street Maps has billions of points needed just to build the map.
// For example, in one of the tests, the first point inserted/searched on the map has a response time in the order of
// nanoseconds, but, after a few billion points, the insertion/search time increased to 500ms per point
// inserted/retrieved in the database, making the project unfeasible.
//
// Português:
//
// A busca binária é usada para popular o banco de dados por causa da curva de tempo de resposta do banco de dados.
//
// O problema: a medida que o banco de dados vai sendo populado com os nodes apenas de construção, o tempo de
// inserção/busca no banco de dados sobe muito e o Open Street Maps tem bilhões de pontos necessários apenas para
// construção do mapa.
// Por exemplo, em um dos testes, o primeiro ponto inserido/buscado no mapa tem um tempo de resposta na casa de
// nanosegundos, mas, depois de alguns bilhões de pontos, o tempo de inserção/busca subiu para 500ms por ponto inserido
// ou recuperado no banco, inviabilizando o projeto.
type PbfProcess struct {
	compress CompressInterface

	totalOfNodesInTmpFile uint64
	totalOfWaysInTmpFile  uint64
	downloadApi           InterfaceDownloadOsm
	database              InterfaceDatabase
	databaseTimeout       time.Duration
}

// SetDatabase
//
// English:
//
// Defines the database control object.
//
// Português:
//
// Define o objeto de controle do banco de dados.
func (e *PbfProcess) SetDatabase(database InterfaceDatabase) {
	e.database = database
}

// SetDatabaseTimeout
//
// English:
//
// # Sets the maximum time for database operations
//
// Português:
//
// Define o tempo máximo para as operações de banco de dados
func (e *PbfProcess) SetDatabaseTimeout(timeout time.Duration) {
	e.databaseTimeout = timeout
}

// SetDownloadApi
//
// English:
//
// Defines the object responsible for downloading information not found in the binary file.
//
// Português:
//
// Define o objeto responsável por download de informações não encontradas no arquivo binário.
func (e *PbfProcess) SetDownloadApi(downloadApi InterfaceDownloadOsm) {
	e.downloadApi = downloadApi
}

func (e *PbfProcess) SetCompress(compress CompressInterface) {
	e.compress = compress
}

// WaysToDatabase
//
// English:
//
// Process the Open Street Maps binary file and populate the database.
//
//	Input:
//	  osmFilePath: Open Street Maps binary file path;
//	  tmpFilePath: path of the binary file generated by the SaveNodesIntoTmpFile() function.
//
//	Notes:
//	  * Before using this function it is necessary to use the SetDatabase(), SetDatabaseTimeout() and SetDownloadApi()
//	    functions
//
// Português:
//
// Processa o arquivo binário do Open Street Maps e popula o banco de dados.
//
//	Entrada:
//	  osmFilePath: caminho do arquivo binário do Open Street Maps;
//	  tmpFilePath: caminho do arquivo binário gerado pela função SaveNodesIntoTmpFile().
//
//	Notas:
//	  * Antes de usar esta função é necessário usar as funções SetDatabase(), SetDatabaseTimeout() e SetDownloadApi()
func (e *PbfProcess) WaysToDatabase(osmFilePath string) (err error) {
	if e.compress == nil {
		err = errors.New("compress must be set")
		return
	}

	if e.database == nil {
		err = errors.New("define a database object before use this function")
		return
	}

	if e.downloadApi == nil {
		err = errors.New("define a download api object before use this function")
		return
	}

	if e.databaseTimeout == 0 {
		err = errors.New("define a database timeout before use this function")
		return
	}

	err = e.compress.ReadFileHeaders()
	if err != nil {
		return
	}

	err = e.compress.IndexToMemory()
	if err != nil {
		return
	}

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

	lon := 0.0
	lat := 0.0
	wayList := make([]Way, 0)
	counter := 0

	tmpNode := Node{}

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch converted := osmPbfElement.(type) {

			case *osmpbf.Way:
				if converted.Info.Visible == false {
					continue
				}

				way := Way{}
				way.Id = converted.ID
				way.Loc = make([][2]float64, len(converted.NodeIDs))
				//way.IdList = converted.NodeIDs
				way.Tag = converted.Tags

				for nodeKey, nodeID := range converted.NodeIDs {
					lon, lat, err = e.compress.FindNodeByID(nodeID)

					// English: downloads points not present in binary file
					// Português: faz o download de pontos não presentes no arquivo binário
					if err != nil && err == io.EOF {
						log.Printf("download ID: %v", nodeID)
						tmpNode, err = e.downloadApi.DownloadNode(nodeID)
						if err != nil {
							return
						}
						lon = tmpNode.Loc[Longitude]
						lat = tmpNode.Loc[Latitude]
					}

					if err != nil {
						return
					}

					way.Loc[nodeKey] = [2]float64{lon, lat}
				}

				err = way.Init()
				if err != nil {
					return
				}
				way.MakeGeoJSonFeature()

				wayList = append(wayList, way)
				counter++

				if counter == 100 {
					err = e.database.SetWay(&wayList, e.databaseTimeout)
					// todo: em caso de erro, inserir um por um e devolver os ways com erro
					if err != nil {
						return
					}
					counter = 0
					wayList = make([]Way, 0)
				}

			case *osmpbf.Relation:

				err = e.database.SetWay(&wayList, e.databaseTimeout)
				return

			}
		}
	}

	return
}

// NodesToDatabase
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
func (e *PbfProcess) NodesToDatabase(osmFilePath string) (err error) {
	if e.database == nil {
		err = errors.New("define a database object before use this function")
		return
	}

	if e.downloadApi == nil {
		err = errors.New("define a download api object before use this function")
		return
	}

	if e.databaseTimeout == 0 {
		err = errors.New("define a database timeout before use this function")
		return
	}

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

	var nodeList = make([]Node, 0)

	var nc uint64
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
						node := Node{}
						node.Init(converted.ID, converted.Lon, converted.Lat, &converted.Tags)
						node.MakeGeoJSonFeature()
						if len(node.Tag) == 0 {
							continue
						}

						nodeList = append(nodeList, node)
						if len(nodeList) == 100 {
							err = e.database.SetNodeMany(&nodeList)
							if err != nil {
								return
							}

							nodeList = make([]Node, 0)
						}
					}
				}

			case *osmpbf.Way:

				if nodeList != nil && len(nodeList) != 0 {
					err = e.database.SetNodeMany(&nodeList)
					if err != nil {
						return
					}

					nodeList = nil
				}

				return
			}
		}
	}

	return
}

// binarySearchInNodeFile
//
// English:
//
// # This function is called by the nodeSearchInTmpFile() function is a variation of golang's sort.Search() function
//
// Português:
//
// Esta função é chamada pela função nodeSearchInTmpFile() é uma variação da função sort.Search() do golang
func (e *PbfProcess) binarySearchInNodeFile(f *os.File, length, value uint64, function func(f *os.File, index, value uint64) (found bool, err error)) (index int64, err error) {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	found := false
	i, j := uint64(0), length
	for i < j {
		h := i + j>>1 // avoid overflow when computing h
		// i ≤ h < j
		found, err = function(f, h, value)
		if err != nil {
			return -1, err
		}
		if !found {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return int64(i), nil
}

// NodeSearchInTmpFile
//
// English:
//
// Does a binary search on the file created by the SaveNodesIntoTmpFile() function.
//
//	Input:
//	  nodeId: id of the node contained in the open street maps file;
//	  tmpFilePath: path to the file generated by the SaveNodesIntoTmpFile() function;
//	  totalOfNodes: total number of nodes calculated by the OsmCounter() function
//
//	Output:
//	  found: node id found in file
//	  loc: [2]float{longitude, latitude}
//	  err: golang standard error object
//
//	Notes:
//	  * float64 is limited to 7 decimal places to avoid the algorithm's least significant value problem, where 1.0 can
//	    be 0.99999999999999
//
// Português:
//
// Faz uma busca binária no arquivo criado pela função SaveNodesIntoTmpFile().
//
//	Entrada:
//	  nodeId: id do node contido no arquivo do open street maps;
//	  tmpFilePath: caminho para o arquivo gerado pela função SaveNodesIntoTmpFile();
//	  totalOfNodes: quantidade total de nodes calculado pela função OsmCounter()
//
//	Saída:
//	  found: id do node encontrado no arquivo
//	  loc: [2]float{longitude, latitude}
//	  err: objeto de erro padrão do golang
//
//	Notas:
//	  * float64 é limitado para 7 casas decimais para evitar o problema do valor menos significativo do algoritmo,
//	    onde 1.0 pode ser 0.99999999999999
func (e *PbfProcess) _nodeSearchInTmpFile(tmpFile *os.File, totalOfNodes, nodeId int64) (found bool, loc [2]float64, err error) {

	fileBinarySearch := func(f *os.File, index, value uint64) (found bool, err error) {
		data := make([]byte, 8)
		_, err = f.ReadAt(data, int64(index*8*3))
		valueRead := binary.LittleEndian.Uint64(data)
		found = valueRead >= value
		return
	}

	var i int64
	i, err = e.binarySearchInNodeFile(tmpFile, uint64(totalOfNodes), uint64(nodeId), fileBinarySearch)
	if err != nil {
		return
	}

	data := make([]byte, 8)
	_, err = tmpFile.ReadAt(data, i*(8+8+8)+0) //id size + lon size + lat size = +0 = id
	if err != nil {
		return
	}

	valueRead := binary.LittleEndian.Uint64(data)

	if i < totalOfNodes && valueRead == uint64(nodeId) {
		found = true
		_, err = tmpFile.ReadAt(data, i*(8+8+8)+8) //id size + lon size + lat size = +8 = longitude
		if err != nil {
			return
		}

		v := binary.LittleEndian.Uint64(data)
		loc[0] = util.Round(math.Float64frombits(v))

		_, err = tmpFile.ReadAt(data, i*(8+8+8)+8+8) //id size + lon size + lat size = +16 = latitude
		if err != nil {
			return
		}

		v = binary.LittleEndian.Uint64(data)
		loc[1] = util.Round(math.Float64frombits(v))
	}

	return
}

// Counter
//
// English:
//
// It counts the amount of elements contained in the binary file.
//
// Português:
//
// Conta a quantidade de elementos contidos no arquivo binário.
func (e *PbfProcess) Counter(osmFilePath string) (nc, wc, rc uint64, err error) {

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

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		} else {
			switch osmPbfElement.(type) {
			case *osmpbf.Node:
				nc++

			case *osmpbf.Way:
				wc++

			case *osmpbf.Relation:
				rc++
			}
		}
	}

	return
}

// SaveNodesIntoTmpFile
//
// English:
//
// # Saves the temporary file for the binary search done by the WaysToDatabase() function
//
// Português:
//
// Salva o arquivo temporário para a busca binária feita pela função WaysToDatabase()
func (e *PbfProcess) SaveNodesIntoTmpFile(osmFilePath string) (nodes uint64, err error) {

	if e.compress == nil {
		err = errors.New("compress must be set")
		return
	}

	e.totalOfNodesInTmpFile = 0

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		err = fmt.Errorf("SaveNodesIntoTmpFile().Open().Error: %v", err)
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
		err = fmt.Errorf("SaveNodesIntoTmpFile().Start().Error: %v", err)
		return
	}

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("SaveNodesIntoTmpFile().Decode().Error: %v", err)
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:

				e.totalOfNodesInTmpFile++

				err = e.compress.WriteNode(converted.ID, converted.Lon, converted.Lat)
				if err != nil {
					err = fmt.Errorf("SaveNodesIntoTmpFile().WriteNode().Error: %v", err)
					return
				}

			default:
				break
			}
		}
	}

	err = e.compress.WriteFileHeaders()
	if err != nil {
		err = fmt.Errorf("SaveNodesIntoTmpFile().WriteFileHeaders().Error: %v", err)
		return
	}

	err = e.compress.MountIndexIntoFile()
	if err != nil {
		err = fmt.Errorf("SaveNodesIntoTmpFile().MountIndexIntoFile().Error: %v", err)
		return
	}

	nodes = e.totalOfNodesInTmpFile
	return
}

// CompleteParser
//
// English:
//
// Processes the open street maps file and inserts all the data found in the data source in an optimized way for the
// planetary file.
//
// Português:
//
// Faz o processamento do arquivo do open street maps e insere todos os dados encontrados na fonte de dados e forma
// otimizada para o arquivo planetário.
func (e *PbfProcess) CompleteParser(osmFilePath string) (nodes, ways uint64, err error) {

	if e.compress == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the compression object must be defined before this function is called")
		return
	}

	if e.downloadApi == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the download object must be defined before this function is called")
		return
	}

	if e.database == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the database object must be defined before this function is called")
		return
	}

	e.totalOfNodesInTmpFile = 0
	e.totalOfWaysInTmpFile = 0

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		err = fmt.Errorf("PbfProcess.CompleteParser().Open().Error: %v", err)
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
		err = fmt.Errorf("PbfProcess.CompleteParser().Start().Error: %v", err)
		return
	}

	writeHeaders := true
	nodeList := make([]Node, 0)
	lon := 0.0
	lat := 0.0
	wayList := make([]Way, 0)
	counter := 0

	tmpNode := Node{}

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("PbfProcess.CompleteParser().Decode().Error: %v", err)
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:

				e.totalOfNodesInTmpFile++

				err = e.compress.WriteNode(converted.ID, converted.Lon, converted.Lat)
				if err != nil {
					err = fmt.Errorf("PbfProcess.CompleteParser()WriteNode().Error: %v", err)
					return
				}

				if converted.Info.Visible && len(converted.Tags) != 0 {

					if len(converted.Tags) != 0 {
						node := Node{}
						node.Init(converted.ID, converted.Lon, converted.Lat, &converted.Tags)
						node.MakeGeoJSonFeature()
						if len(node.Tag) == 0 {
							continue
						}

						nodeList = append(nodeList, node)
						if len(nodeList) == 100 {
							err = e.database.SetNodeMany(&nodeList)
							if err != nil {
								return
							}

							nodeList = make([]Node, 0)
						}
					}
				}

			case *osmpbf.Way:

				e.totalOfWaysInTmpFile++

				if writeHeaders == true {
					writeHeaders = false

					err = e.compress.WriteFileHeaders()
					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().WriteFileHeaders().Error: %v", err)
						return
					}

					err = e.compress.MountIndexIntoFile()
					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().MountIndexIntoFile().Error: %v", err)
						return
					}

					err = e.compress.ReadFileHeaders()
					if err != nil {
						return
					}

					err = e.compress.IndexToMemory()
					if err != nil {
						return
					}
				}

				if nodeList != nil && len(nodeList) != 0 {
					err = e.database.SetNodeMany(&nodeList)
					if err != nil {
						return
					}

					nodeList = nil
				}

				if converted.Info.Visible == false {
					continue
				}

				way := Way{}
				way.Id = converted.ID
				way.Loc = make([][2]float64, len(converted.NodeIDs))
				//way.IdList = converted.NodeIDs
				way.Tag = converted.Tags

				for nodeKey, nodeID := range converted.NodeIDs {
					lon, lat, err = e.compress.FindNodeByID(nodeID)

					// English: downloads points not present in binary file
					// Português: faz o download de pontos não presentes no arquivo binário
					if err != nil && err == io.EOF {
						log.Printf("PbfProcess.CompleteParser().event: download ID: %v", nodeID)
						tmpNode, err = e.downloadApi.DownloadNode(nodeID)
						if err != nil {
							return
						}
						lon = tmpNode.Loc[Longitude]
						lat = tmpNode.Loc[Latitude]
					}

					if err != nil {
						return
					}

					way.Loc[nodeKey] = [2]float64{lon, lat}
				}

				err = way.Init()
				if err != nil {
					return
				}
				way.MakeGeoJSonFeature()

				wayList = append(wayList, way)
				counter++

				if counter == 100 {
					err = e.database.SetWay(&wayList, e.databaseTimeout)
					// todo: em caso de erro, inserir um por um e devolver os ways com erro
					if err != nil {
						return
					}
					counter = 0
					wayList = make([]Way, 0)
				}

			case *osmpbf.Relation:

				err = e.database.SetWay(&wayList, e.databaseTimeout)
				return

			default:
				err = errors.New("PbfProcess.CompleteParser().error: formato de dado não previsto no arquivo pbf do open street maps")
				return
			}
		}
	}

	ways = e.totalOfWaysInTmpFile
	nodes = e.totalOfNodesInTmpFile
	return
}

func (e *PbfProcess) WrongWayParser(osmFilePath, timeLogPath string) (err error) {

	if e.database == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the database object must be defined before this function is called")
		return
	}

	e.totalOfNodesInTmpFile = 0
	e.totalOfWaysInTmpFile = 0

	var timeLog *os.File
	timeLog, err = os.OpenFile(timeLogPath, os.O_CREATE|os.O_WRONLY, fs.ModePerm)
	if err != nil {
		err = fmt.Errorf("PbfProcess.CompleteParser().Open(timeLog).Error: %v", err)
		return
	}

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		err = fmt.Errorf("PbfProcess.CompleteParser().Open(osm).Error: %v", err)
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
		err = fmt.Errorf("PbfProcess.CompleteParser().Start().Error: %v", err)
		return
	}

	nodeList := make([]Node, 0)
	node := Node{}

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("PbfProcess.CompleteParser().Decode().Error: %v", err)
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:

				e.totalOfNodesInTmpFile++

				node.Init(converted.ID, converted.Lon, converted.Lat, &converted.Tags)
				nodeList = append(nodeList, node)
				if len(nodeList) == 100 {
					start := time.Now()
					err = e.database.SetWrongWayNodeMany(&nodeList)
					if err != nil {
						err = fmt.Errorf("WrongWayParser().error: the function Osm.SetNodeMany() returned an error: %v", err)
						return
					}
					duration := time.Since(start)
					_, err = timeLog.WriteString(strconv.FormatInt(int64(e.totalOfNodesInTmpFile), 10))
					if err != nil {
						err = fmt.Errorf("WrongWayParser().error: the function timeLog.WriteString() returned an error: %v", err)
						return
					}
					_, err = timeLog.WriteString(",")
					if err != nil {
						err = fmt.Errorf("WrongWayParser().error: the function timeLog.WriteString() returned an error: %v", err)
						return
					}
					_, err = timeLog.WriteString(strconv.FormatInt(duration.Microseconds(), 10) + "\r\n")
					if err != nil {
						err = fmt.Errorf("WrongWayParser().error: the function timeLog.WriteString() returned an error: %v", err)
						return
					}

					nodeList = make([]Node, 0)
				}

			case *osmpbf.Way:
				return

			case *osmpbf.Relation:
				return

			default:
				err = errors.New("PbfProcess.CompleteParser().error: formato de dado não previsto no arquivo pbf do open street maps")
				return
			}
		}
	}

	return
}

// GetPartialNumberOfProcessedData
//
// English:
//
// # Returns the partial amount of processed data
//
// Português:
//
// Retorna a quantidade parcial de dados processados
func (e *PbfProcess) GetPartialNumberOfProcessedData() (nodes uint64, ways uint64) {
	nodes = e.totalOfNodesInTmpFile
	ways = e.totalOfWaysInTmpFile

	return
}

// binarySearchList
//
// English:
//
// Português:
//
// Retorna a chave onde o ID se encaixa na memória, ou seja, o ID procurado se encontra entre os endereços
// (key * block) e ((key + 1) * block).
//
// Caso `value` seja diferente de zero, (key * block) é o local do dado procurado.
func (e *PbfProcess) binarySearchList(memory *[]int64, ID int64) (key int, value int64) {
	key = sort.Search(len(*memory), func(i int) bool { return (*memory)[i] >= ID })
	if key < len(*memory) && (*memory)[key] == ID {
		value = (*memory)[key]
		return
	}

	if key > 0 {
		key--
	}

	return
}

// countNodesIntoTmpFile
//
// English:
//
// Counts the total number of nodes saved in the temporary file for the binary search.
//
// Português:
//
// Conta a quantidade total de nodes salvos no arquivo temporário para a busca binária.
func (e *PbfProcess) countNodesIntoTmpFile(tmpFilePath string) (err error) {
	e.totalOfNodesInTmpFile = 0

	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_RDONLY, fs.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	var filePointer = int64(0)
	var idAsByte = make([]byte, 8)

	for {
		_, err = tmpFile.ReadAt(idAsByte, filePointer)
		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		// English: This counter has to stay after EOF check
		// Português: Este contador tem que ficar depois da verificação de EOF
		e.totalOfNodesInTmpFile++

		// 8 bytes, node ID + 8 bytes, longitude + 8 bytes latitude
		filePointer += 8 + 8 + 8
	}
}
