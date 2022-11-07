package goosm

import (
	"errors"
	"fmt"
	"github.com/qedus/osmpbf"
	"io"
	"log"
	"os"
	"runtime"
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

	// DownloadNode
	//
	// English:
	//
	// # Download the initialized node for use
	//
	// Português:
	//
	// Faz o download do node inicializado para uso
	DownloadNode(id int64) (node Node, err error)

	// DownloadWay
	//
	// English:
	//
	// Downloads the initialized way for use.
	//
	// Português:
	//
	// Faz o download do way inicializado para uso.
	DownloadWay(id int64) (way Way, err error)

	// DownloadRelation
	//
	// English:
	//
	// Downloads the initialized relation for use.
	//
	// Português:
	//
	// Faz o download da relation inicializado para uso.
	DownloadRelation(id int64) (relation Relation, err error)
}

type InterfaceConnect interface {
	// SetTimeout
	//
	// English:
	//
	// Determines timeout for all functions
	//
	//  Input:
	//    timeout: maximum time for operation
	//
	// Português:
	//
	// Determina o timeout para todas as funções
	//
	//  Entrada:
	//    timeout: tempo máximo para a operação
	SetTimeout(timeout time.Duration)

	// Connect
	//
	// English:
	//
	// Connect to the database
	//
	//  Input:
	//    connection: database connection string. eg. "mongodb://127.0.0.1:27016/"
	//    args: maintained by interface compatibility
	//
	// Português:
	//
	// Conecta ao banco de dados
	//
	//  Entrada:
	//    connection: string de conexão ao banco de dados. Ex: "mongodb://127.0.0.1:27016/"
	//    args: mantido por compatibilidade da interface
	Connect(connection string, _ ...interface{}) (err error)

	// Close
	//
	// English:
	//
	// Close the connection to the database
	//
	// Português:
	//
	// Fecha a conexão com o banco de dados
	Close() (err error)

	// New
	//
	// English:
	//
	// Prepare the database for use
	//
	//  Input:
	//    connection: database connection string. Eg: "mongodb://127.0.0.1:27016/"
	//    database: database name. Eg. "osm"
	//    collection: collection name within the database. Eg. "way"
	//
	//  Output:
	//    referenceInitialized: database way object ready to use
	//    err: golang error object
	//
	// Português:
	//
	// Prepara o banco de dados para uso
	//
	//  Entrada:
	//    connection: string de conexão ao banco de dados. Ex: "mongodb://127.0.0.1:27016/"
	//    database: nome do banco de dados. Ex: "osm"
	//    collection: nome da coleção dentro do banco de dados. Ex: "way"
	//
	//  Saída:
	//    referenceInitialized: objeto do banco de dados pronto para uso
	//    err: objeto golang error
	New(connection, database, collection string) (referenceInitialized interface{}, err error)
}

type InterfaceDbWay interface {
	// SetOne
	//
	// English:
	//
	// Insert a single way into the database
	//
	//  Input:
	//    way: reference to object goosm.Way
	//
	// Português:
	//
	// Insere um único way no banco de dados
	//
	//  Entrada:
	//    way: referencia ao objeto goosm.Way.
	SetOne(way *Way) (err error)

	// GetById
	//
	// English:
	//
	// Returns a way according to ID
	//
	//  Input:
	//    id: ID in the Open Street Maps project pattern
	//
	// Português:
	//
	// Retorna um way de acordo com o ID
	//
	//  Entrada:
	//    id: ID no padrão do projeto Open Street Maps.
	GetById(id int64) (way Way, err error)

	// SetMany
	//
	// English:
	//
	// Insert a block of ways into the database
	//
	//  Input:
	//    list: reference to slice with []goosm.Way objects
	//
	// Português:
	//
	// Insere um bloco de ways no banco de dados
	//
	//  Entrada:
	//    list: referência ao slice com os objetos []goosm.Way
	SetMany(list *[]Way) (err error)
}

type InterfaceDbNode interface {
	// SetOne
	//
	// English:
	//
	// Insert a single node into the database
	//
	//  Input:
	//    node: reference to object goosm.Node
	//
	// Português:
	//
	// Insere um único node no banco de dados
	//
	//  Entrada:
	//    node: referencia ao objeto goosm.Node
	SetOne(node *Node) (err error)

	// GetById
	//
	// English:
	//
	// Returns a node according to ID
	//
	//  Input:
	//    id: ID in the Open Street Maps project pattern
	//
	// Português:
	//
	// Retorna um node de acordo com o ID
	//
	//  Entrada:
	//    id: ID no padrão do projeto Open Street Maps
	GetById(id int64) (node Node, err error)

	// SetMany
	//
	// English:
	//
	// Insert a block of nodes into the database
	//
	//  Input:
	//    list: reference to slice with []goosm.Node objects
	//
	// Português:
	//
	// Insere um bloco de nodes no banco de dados
	//
	//  Entrada:
	//    list: referência ao slice com os objetos []goosm.Node
	SetMany(list *[]Node) (err error)
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
	databaseNode          InterfaceDbNode
	databaseWay           InterfaceDbWay
	databaseTimeout       time.Duration
}

// SetDatabaseNode
//
// English:
//
// # Defines the object for inserting nodes into the database
//
// Português:
//
// Define o objeto de inserção de nodes no banco de dados
func (e *PbfProcess) SetDatabaseNode(database InterfaceDbNode) {
	e.databaseNode = database
}

// SetDatabaseWay
//
// English:
//
// # Defines the insertion object of ways in the database
//
// Português:
//
// Define o objeto de inserção de ways no banco de dados
func (e *PbfProcess) SetDatabaseWay(database InterfaceDbWay) {
	e.databaseWay = database
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

// SetCompress
//
// English:
//
// Defines the compression object responsible for the non-database binary search.
//
// Português:
//
// Define o objeto de compressão responsável pela busca binária, sem banco de dados.
func (e *PbfProcess) SetCompress(compress CompressInterface) {
	e.compress = compress
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

	if e.databaseNode == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the databaseNode object must be defined before this function is called")
		return
	}

	if e.databaseWay == nil {
		err = errors.New("PbfProcess.CompleteParser().error: the databaseWay object must be defined before this function is called")
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
							err = e.databaseNode.SetMany(&nodeList)
							if err != nil {
								err = fmt.Errorf("PbfProcess.CompleteParser().SetMany().Error: %v", err)
								return
							}

							nodeList = make([]Node, 0)
						}
					}
				}

			case *osmpbf.Way:

				e.totalOfWaysInTmpFile++

				if writeHeaders {
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
						err = fmt.Errorf("PbfProcess.CompleteParser().ReadFileHeaders().Error: %v", err)
						return
					}

					err = e.compress.IndexToMemory()
					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().IndexToMemory().Error: %v", err)
						return
					}
				}

				// English: The amount of data in the planetary file is very large and comparing with nil is faster.
				// Português: A quantidade de dados no arquivo planetário é muito grande e comparar com nil é mais rápido.
				if nodeList != nil && len(nodeList) != 0 { //nolint:typecheck
					err = e.databaseNode.SetMany(&nodeList)
					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().SetMany().Error: %v", err)
						return
					}

					nodeList = nil
				}

				if !converted.Info.Visible {
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
							err = fmt.Errorf("PbfProcess.CompleteParser().DownloadNode().Error: %v", err)
							return
						}
						lon = tmpNode.Loc[Longitude]
						lat = tmpNode.Loc[Latitude]
					}

					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().FindNodeByID().Error: %v", err)
						return
					}

					way.Loc[nodeKey] = [2]float64{lon, lat}
				}

				err = way.Init()
				if err != nil {
					err = fmt.Errorf("PbfProcess.CompleteParser().Init().Error: %v", err)
					return
				}
				way.MakeGeoJSonFeature()

				wayList = append(wayList, way)
				counter++

				if counter == 100 {
					err = e.databaseWay.SetMany(&wayList)
					// todo: em caso de erro, inserir um por um e devolver os ways com erro
					if err != nil {
						err = fmt.Errorf("PbfProcess.CompleteParser().SetMany(1).Error: %v", err)
						return
					}
					counter = 0
					wayList = make([]Way, 0)
				}

			case *osmpbf.Relation:

				err = e.databaseWay.SetMany(&wayList)
				err = fmt.Errorf("PbfProcess.CompleteParser().SetMany(2).Error: %v", err)
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

// BinaryNodeOnlyParser
//
// English:
//
// Processes the open street maps file and make only the binary file.
//
// Português:
//
// Faz o processamento do arquivo do open street maps e faz apenas o arquivo binário.
func (e *PbfProcess) BinaryNodeOnlyParser(osmFilePath string) (nodes, ways uint64, err error) {

	if e.compress == nil {
		err = errors.New("PbfProcess.BinaryNodeOnlyParser().error: the compression object must be defined before this function is called")
		return
	}

	if e.downloadApi == nil {
		err = errors.New("PbfProcess.BinaryNodeOnlyParser().error: the download object must be defined before this function is called")
		return
	}

	e.totalOfNodesInTmpFile = 0
	e.totalOfWaysInTmpFile = 0

	var osmFile *os.File
	osmFile, err = os.Open(osmFilePath)
	if err != nil {
		err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().Open().Error: %v", err)
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
		err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().Start().Error: %v", err)
		return
	}

	for {
		var osmPbfElement interface{}
		if osmPbfElement, err = osmDecoder.Decode(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().Decode().Error: %v", err)
			return
		} else {
			switch converted := osmPbfElement.(type) {
			case *osmpbf.Node:

				e.totalOfNodesInTmpFile++

				err = e.compress.WriteNode(converted.ID, converted.Lon, converted.Lat)
				if err != nil {
					err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser()WriteNode().Error: %v", err)
					return
				}

			case *osmpbf.Way:

				err = e.compress.WriteFileHeaders()
				if err != nil {
					err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().WriteFileHeaders().Error: %v", err)
					return
				}

				err = e.compress.MountIndexIntoFile()
				if err != nil {
					err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().MountIndexIntoFile().Error: %v", err)
					return
				}

				err = e.compress.ReadFileHeaders()
				if err != nil {
					err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().ReadFileHeaders().Error: %v", err)
					return
				}

				err = e.compress.IndexToMemory()
				if err != nil {
					err = fmt.Errorf("PbfProcess.BinaryNodeOnlyParser().IndexToMemory().Error: %v", err)
					return
				}

				return

			default:
				err = errors.New("PbfProcess.BinaryNodeOnlyParser().error: formato de dado não previsto no arquivo pbf do open street maps")
				return
			}
		}
	}

	ways = e.totalOfWaysInTmpFile
	nodes = e.totalOfNodesInTmpFile
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
