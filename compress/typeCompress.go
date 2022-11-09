package compress

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"sort"
)

const (

	// decimalPlaces
	//
	// English:
	//
	// Used to multiply and divide longitude and latitude during floating point to integer transformation to binary
	// conversion.
	//
	// Português:
	//
	// Usado para multiplicar e dividir longitude e latitude durante a transformação de ponto flutuante em inteiro para a
	// conversão de binário.
	decimalPlaces = 10000000.0

	// headerVersionAddress
	//
	// English:
	//
	// Binary file version header address
	//
	// Português:
	//
	// Endereço do cabeçalho de versão do arquivo binário
	headerVersionAddress = 0

	// headerTotalNodesAddress
	//
	// English:
	//
	// Header address of the total nodes contained in the binary file
	//
	// Português:
	//
	// Endereço do cabeçalho do total de nodes contidos no arquivo binário
	headerTotalNodesAddress = 8

	// headerBlockSizeAddress
	//
	// English:
	//
	// Header address containing spacing size between address blocks in binary search
	//
	// Português:
	//
	// Endereço do cabeçalho contendo o tamanho do espaçamento entre os blocos de endereço na busca binária
	headerBlockSizeAddress = 8 + 8

	// headerTotalIndexAddress
	//
	// English:
	//
	// Header address with total number of indices for secondary binary search
	//
	// Português:
	//
	// Endereço do cabeçalho com a quantidade total de índices para a busca binária secundária
	headerTotalIndexAddress = 8 + 8 + 8

	// headerIndexesPositionAddress
	//
	// English:
	//
	// Address of start of node data
	//
	// Português:
	//
	// Endereço do início dos dados de nodes
	headerIndexesPositionAddress = 8 + 8 + 8 + 8

	// headerVersion
	//
	// English:
	//
	// Version text written in binary file header
	//
	// Português:
	//
	// Texto de versão escrito no cabeçalho do arquivo binário
	headerVersion = "00000001"

	// headerVersionByteSize
	//
	// English:
	//
	// Number of letters counted in the version text written in the header of the binary file
	//
	// Português:
	//
	// Quantidade de letras contadas no texto de versão escrito no cabeçalho do arquivo binário
	headerVersionByteSize = 8

	// int64ByteSize
	//
	// English:
	//
	// Number of bytes occupied by a 64-bit integer type number
	//
	// Português:
	//
	// Quantidade de bytes ocupada por um número do tipo inteiro de 64 bits
	int64ByteSize = 8

	// totalNodesByteSize
	//
	// English:
	//
	// Number of bytes occupied by the information, total nodes contained in the binary file
	//
	// Português:
	//
	// Quantidade de bytes ocupada pela informação, total de nodes contidos no arquivo binário
	totalNodesByteSize = int64ByteSize

	// totalBlockSizeByteSize
	//
	// English:
	//
	// Number of bytes occupied by the information, secondary binary search block size
	//
	// Português:
	//
	// Quantidade de bytes ocupada pela informação, tamanho do bloco da busca binária secundária
	totalBlockSizeByteSize = int64ByteSize

	// totalIndexIntoFileByteSize
	//
	// English:
	//
	// Number of bytes occupied by the information, total number of indexes contained in the binary file
	//
	// Português:
	//
	// Quantidade de bytes ocupada pela informação, quantidade total de índices contido no arquivo binário
	totalIndexIntoFileByteSize = int64ByteSize

	// startIndexAddress
	//
	// English:
	//
	// Number of bytes occupied by the information, starting address of secondary indexes
	//
	// Português:
	//
	// Quantidade de bytes ocupada pela informação, endereço inicial dos índices secundários
	startIndexAddress = int64ByteSize

	// nodeDataPositionStartAtAddress
	//
	// English:
	//
	// Total amount of bytes to be skipped before writing node data
	//
	// Português:
	//
	// Quantidade total de bytes a ser pulada antes de escrever os dados dos nodes
	nodeDataPositionStartAtAddress = headerVersionByteSize + totalNodesByteSize + totalBlockSizeByteSize + totalIndexIntoFileByteSize + startIndexAddress

	// nodeIdByteSize
	//
	// English:
	//
	// Number of bytes occupied by the node ID
	//
	// Português:
	//
	// Quantidade de bytes ocupada pelo ID do node
	nodeIdByteSize = int64ByteSize

	// nodeCoordinateByteSize
	//
	// English:
	//
	// Number of bytes occupied by a geographic coordinate
	//
	// Português:
	//
	// Quantidade de bytes ocupada por uma coordenada geográfica
	nodeCoordinateByteSize = 4

	// mostSignificantBit
	//
	// English:
	//
	// Position of the most significant bit when little endian is chosen when converting to binary
	//
	// Português:
	//
	// Posição do bit mais significativo quando é escolhido little endian na conversão para binário
	mostSignificantBit = 0x80

	// mostSignificantBitTwoComplements
	//
	// English:
	//
	// Separated from the most significant bit
	//
	// Português:
	//
	// Separado do bit mais significativo
	mostSignificantBitTwoComplements = 0x7F

	// mostSignificantByte
	//
	// English:
	//
	// Most significant byte in coordinate byte set (4 bytes), when little endian of binary is used
	//
	// Português:
	//
	// Byte mais significativo no conjunto de bytes da coordenada (4 bytes), quando é usado little endian do binário
	mostSignificantByte = 3

	// nodeDataByteSize
	//
	// English:
	//
	// Total number of bytes occupied by a node
	//
	// Português:
	//
	// Quantidade total de bytes ocupado por um node
	nodeDataByteSize = nodeIdByteSize + 2*nodeCoordinateByteSize

	// memorySliceAddrID
	//
	// English:
	//
	// ID address within the memory slice
	//
	// Português:
	//
	// Endereço do ID dentro do slice de memória
	memorySliceAddrID = 0

	// memorySliceAddrOfAddrIntoFile
	//
	// English:
	//
	// Address, within the memory slice, containing the address of the data in the nodes file
	//
	// Português:
	//
	// Endereço, dentro do slice de memória, contendo o endereço do dado no arquivo de nodes
	memorySliceAddrOfAddrIntoFile = 1
)

// Compress
//
// English:
//
// File format:
//
//	Header: 24 bytes
//	  total of nodes in a file: 8 bytes
//	  total block size: 8 bytes
//	  start index address: 8 bytes
//
//	Data block:
//	  node.ID: 8 bytes
//	  node.Longitude: 4 bytes
//	  node.Latitude: 4 bytes
//
//	  The largest number present in a coordinate for the flat map is +-180˚ longitude and +-90˚ latitude, to 7 decimal
//	  places.
//	  For data compression, the floating point coordinate is multiplied by 10,000,000 and then converted to an integer,
//	  losing the decimal part, so the largest number saved is the integer +-1,800,000,000.
//	  Therefore, it can be represented by the group of four bytes X110 1011 0100 1001 1101 0010 0000 0000, where the
//	  most significant bit, `X`, is never used, so it can be used to indicate a positive or negative sign, that is, X=1
//	  represents a negative number and X=0 a positive number.
//
//	Index block:
//	  Indexes are a fixed-size block used for in-memory indexing, where a block represents the node address in the file.
//	  For example:
//	  For the list of node ID 1 to 100, in the file, and block size 10, memory will contain the values 1, 11, 21, ...,
//	  81, 91.
//	  Compress.FindNodeByID(75) will find the values 7 and 8 for the bottom edge and top edge.
//	  left border will have the ID 71, (memory[7][0]), and the address 1.160, (memory[7][1])
//	  right border will have the ID 81, (memory[8][0]) and the address 1.320, (memory[8][1])
//	  Therefore, the ID sought will be between addresses 1160 and 1330; And each address is 8 bytes for ID + 4 bytes
//	  for the longitude + 4 bytes for latitude.
//
// Português:
//
// File format:
//
//	Header: 24 bytes
//	  total of nodes in a file: 8 bytes
//	  total block size: 8 bytes
//	  start index address: 8 bytes
//
//	Data block:
//	  node.ID: 8 bytes
//	  node.Longitude: 4 bytes
//	  node.Latitude: 4 bytes
//
//	  O maior número presente em uma coordenada para o mapa planificado é +/-180˚ na longitude e +/-90˚ na latitude,
//	  com 7 casas decimais.
//	  Para a compactação de dados, o número de ponto flutuante da coordenada é multiplicado por 10.000.000 e em seguida
//	  é convertido em inteiro, perdendo a parte decimal, logo, o maior número salvo é o inteiro +/-1.800.000.000.
//	  Logo, pode ser representado pelo grupo de quatro bytes X110 1011 0100 1001 1101 0010 0000 0000, onde o bit mais
//	  significativo, `X` nunca é usado, logo, pode ser usado para indicar sinal de positivo ou negativo, ou seja, X=1
//	  representa um número negativo e X=0 um número positivo.
//
//	Index block:
//	  Índices são um bloco de tamanho fixo, usado para uma indexação em memória, onde um bloco representa o endereço do
//	  node no arquivo.
//	  Por exemplo:
//	  Para a lista de node ID 1 a 100, no arquivo, e tamanho do bloco 10, memory conterá os valores 1, 11, 21, ...,
//	  81, 91.
//	  Compress.FindNodeByID(75) encontrará os valores 7 e 8 para a borda inferior e a borda superior. Aplicando a fórmula:
//	  a borda esquerda terá o ID 71, (memory[7][0]), e o endereço 1.160, (memory[7][1]),
//	  a borda direita terá o ID 81, (memory[8][0]) e o endereço (memory[8][1]) 1.320
//	  Logo, o ID procurado estará entre os endereços 1.160 e 1320; E cada endereço tem 8 bytes para ID + 4 bytes para
//	  a longitude + 4 bytes para a latitude.
type Compress struct {

	// English:
	//
	// Pointer to the temporary file.
	//
	// Português:
	//
	// Ponteiro para o arquivo temporário.
	file *os.File

	// English:
	//
	// Checks if the entered ID is in ascending order
	//
	// Português:
	//
	// Verifica se o ID inserido esta em ordem crescente
	lastID int64

	// English:
	//
	// 8 bytes to file uint64. Creating this variable here saves time when it comes to more than 8 billion points.
	//
	// Português:
	//
	// 8 bytes to file uint64. Creating this variable here saves time when it comes to more than 8 billion points.
	dataFile []byte

	// English:
	//
	// 4 bytes to archive float64 compression. Creating this variable here saves time when it comes to more than 8
	// billion points.
	//
	// Português:
	//
	// 4 bytes para arquivar a compactação de float64. Criar esta variável aqui ganha tempo quando se trata mais de 8
	// bilhões de pontos.
	dataCoordinate []byte

	// English:
	//
	// Write pointer to temporary file.
	//
	// Português:
	//
	// Ponteiro de escrita no arquivo temporário.
	nodeWriteDataPosition int64

	// English:
	//
	// Read pointer to temporary file.
	//
	// Português:
	//
	// Ponteiro de leitura no arquivo temporário.
	nodeReadDataPosition int64

	// English:
	//
	// Total nodes saved in the file.
	//
	// Português:
	//
	// Total de nodes salvos no arquivo.
	totalOfNodesInTmpFile int64

	// English:
	//
	// Total indexes to be used in memory in the file. Affected by blockSize.
	//
	// Português:
	//
	// Total de índices para serem usados em memória no arquivo. Afetado por blockSize.
	totalIndexIntoFile int64

	// English:
	//
	// Spacing between IDs for the in-memory key, i.e. 10 represents one data capture every 10 IDs.
	// A low value represents greater search efficiency, but greater memory and disk consumption.
	//
	// Português:
	//
	// Espaçamento entre IDs para a chave em memória, ou seja, 10 representa uma captura de dados a cada 10 IDs.
	// Um valor baixo representa uma maior eficiência na busca, porém, um maior consumo de memória e de disco.
	blockSize int64

	// English:
	//
	// Receives the list of node IDs, where (key * block size * (ID size + lon size + lat size) + header size) =
	// = node ID address.
	//
	//  Example:
	//
	//    For the list of node ID 1 to 100, in the file, and block size 10, memory will contain the values 1, 11, 21, ...,
	//    81, 91.
	//    Compress.FindNodeByID(75) will find the values 7 and 8 for the bottom edge and top edge.
	//    left border will have the ID 71, (memory[7][0]), and the address 1.160, (memory[7][1])
	//    right border will have the ID 81, (memory[8][0]) and the address 1.320, (memory[8][1])
	//    Therefore, the ID sought will be between addresses 1160 and 1330; And each address is 8 bytes for ID + 4 bytes
	//    for the longitude + 4 bytes for latitude.
	//
	// Português:
	//
	// Recebe a lista de IDs dos nodes, onde (chave * block size * (ID size + lon size + lat size) + header size) =
	// = node ID address.
	//
	//  Exemplo:
	//
	//    Para a lista de node ID 1 a 100, no arquivo, e tamanho do bloco 10, memory conterá os valores 1, 11, 21, ...,
	//    81, 91.
	//    Compress.FindNodeByID(75) encontrará os valores 7 e 8 para a borda inferior e a borda superior. Aplicando a fórmula:
	//    a borda esquerda terá o ID 71, (memory[7][0]), e o endereço 1.160, (memory[7][1]),
	//    a borda direita terá o ID 81, (memory[8][0]) e o endereço (memory[8][1]) 1.320
	//    Logo, o ID procurado estará entre os endereços 1.160 e 1320; E cada endereço tem 8 bytes para ID + 4 bytes para
	//    a longitude + 4 bytes para a latitude.
	memory [][2]int64
}

// Init
//
// English:
//
// Initializes the object.
//
//	Input:
//	  blockSize: Spacing between ID captures for the in-memory index.
//
//	Note:
//	  * Very small values consume a lot of memory and make the file very large, but generate greater efficiency.
//	  * For the planet file, with ~8 trillions of points, 8000000000 nodes / 1000 block size * (8 bytes per ID + 8 bytes per address) = 122MB
//
// Português:
//
// Inicializa o objeto.
//
//	Entrada:
//	  blockSize: Espaçamento entre as capturas de IDs para o índice em memória.
//
//	Nota:
//	  * Valores muitos pequenos consomem muita memória e deixam o arquivo muito grande, mas, geram maior eficiência.
//	  * Para o arquivo do planeta, com ~8 bilhões de pontos, 8000000000 nodes / 1000 block size * (8 bytes por ID + 8 bytes por endereço) = 122MB
func (e *Compress) Init(blockSize int64) {
	e.dataFile = make([]byte, 8)
	e.dataCoordinate = make([]byte, 4)
	e.nodeWriteDataPosition = nodeDataPositionStartAtAddress
	e.nodeReadDataPosition = nodeDataPositionStartAtAddress
	e.blockSize = blockSize
	e.memory = make([][2]int64, 0)
}

func (e *Compress) OpenForSearch(path string) (err error) {
	if e.file != nil {
		_ = e.file.Close()
	}

	e.file, err = os.OpenFile(path, os.O_APPEND|os.O_RDONLY, fs.ModePerm)
	if err != nil {
		err = fmt.Errorf("Compress.OpenForSearch().OpenFile().Error: %v", err)
		return
	}

	err = e.ReadFileHeaders()
	if err != nil {
		err = fmt.Errorf("Compress.OpenForSearch().ReadFileHeaders().Error: %v", err)
		return
	}

	err = e.IndexToMemory()
	if err != nil {
		err = fmt.Errorf("Compress.OpenForSearch().IndexToMemory().Error: %v", err)
		return
	}

	return
}

// ResizeBlock
//
// English:
//
// # Rewrites block size of indexes and redoes in-memory search indexes
//
// Português:
//
// Reescreve o tamanho do bloco de índices e refaz os índices da busca em memória
func (e *Compress) ResizeBlock(blockSize int64) (err error) {
	err = e.ReadFileHeaders()
	if err != nil {
		err = fmt.Errorf("ResizeBlock().error: the ReadFileHeaders() function returned an error: %v", err)
		return
	}

	e.blockSize = blockSize
	err = e.writeHeaderBlockSize()
	if err != nil {
		err = fmt.Errorf("ResizeBlock().error: the writeHeaderBlockSize() function returned an error: %v", err)
		return
	}

	err = e.writeHeaderTotalIndexIntoFile()
	if err != nil {
		err = fmt.Errorf("ResizeBlock().error: the writeHeaderTotalIndexIntoFile() function returned an error: %v", err)
		return
	}

	err = e.MountIndexIntoFile()
	if err != nil {
		err = fmt.Errorf("ResizeBlock().error: the MountIndexIntoFile() function returned an error: %v", err)
		return
	}

	return
}

// Round
//
// English:
//
// Rounds a floating point to N decimal places
//
//	Input:
//	  value: value to be rounded off;
//	  places: number of decimal places. Eg. 7.0
//
// Português:
//
// Rounds a floating point to N decimal places
//
//	Entrada:
//	  value: valor a ser arredondado;
//	  places: quantidade de casas decimais. Ex: 7.0
func (e *Compress) Round(value, places float64) float64 {
	var roundOn = 0.5

	var round float64
	pow := math.Pow(10, places)
	digit := pow * value
	_, div := math.Modf(digit)

	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return round / pow
}

// Create
//
// English:
//
// Open the temporary file.
//
// Português:
//
// Abre o arquivo temporário.
func (e *Compress) Create(path string) (err error) {
	if e.file != nil {
		_ = e.file.Close()
	}

	e.file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		err = fmt.Errorf("compress.Create().error: the function OpenFile() returned an error: %v", err)
		return
	}
	return
}

// Close
//
// English:
//
// # Close the temporary file
//
// Português:
//
// Fecha o arquivo temporário
func (e *Compress) Close() {
	var err = e.file.Close()
	if err != nil {
		log.Printf("Compress.Close().error: %v", err)
	}
}

// WriteNode
//
// English:
//
// Write node to temporary file.
//
//	Input:
//	  id: positive number greater than zero;
//	  longitude: value between ±180 to 7 decimal places;
//	  latitude: value between ±90 with 7 decimal places;
//
//	Notas:
//	  * Node format saved 8 bytes for ID, 4 bytes for longitude and 4 bytes for latitude;
//	  * Longitude and latitude save 7 decimal places;
//	  * The most significant bit is 1 indicates a negative number, but the two's complement rule was not used.
//	    If the most significant bit is 1, the number is negative, otherwise positive.
//	  * Coordinate compression at 4 bytes saves considerable time and compresses the final file.
//
// Português:
//
// Escreve o node no arquivo temporário.
//
//	Entrada:
//	  id: número positivo maior do que zero;
//	  longitude: valor entre ±180 com 7 casas decimais;
//	  latitude: valor entre ±90 com 7 casas decimais;
//
//	Notes:
//	  * Formato do node salvo 8 bytes para ID, 4 bytes para longitude e 4 bytes para latitude;
//	  * Longitude e latitude salvam 7 casas decimais;
//	  * O bit mais significativo em 1, indica um número negativo, mas, não foi usada a regra do complemento de dois.
//	    Caso o bit mais significativo seja 1, o número é negativo, caso contrário, positivo.
//	  * A compactação de coordenada em 4 bytes gera um ganho de tempo considerável e compacta o arquivo final.
func (e *Compress) WriteNode(id int64, longitude, latitude float64) (err error) {
	if id <= e.lastID {
		err = errors.New("id must be entered in ascending order and must not be repeated")
		return
	}

	if longitude < -180.0 || longitude > 180.0 {
		err = errors.New("longitude must be within ±180˚")
		return
	}

	if latitude < -90.0 || latitude > 90.0 {
		err = errors.New("latitude must be within ±90˚")
		return
	}

	// Why converting integer to byte before archiving
	//
	// | total process | Memory Allocs     | Type of golang data                          |
	// |---------------|-------------------|----------------------------------------------|
	// |  4.397375375s | 1578894104 Allocs | []byte using binary.LittleEndian.PutUint64() |
	// |  3.830812750s | 1511656176 Allocs | []byte using binary.LittleEndian.PutUint64() |
	// |  3.834706875s | 1274415048 Allocs | []byte using binary.LittleEndian.PutUint64() |
	// |---------------|-------------------|----------------------------------------------|
	// | 22.857736375s | 4353555600 Allocs | map[id][]float64{longitude, latitude}        |
	// | 23.104297458s | 4426114960 Allocs | map[id][]float64{longitude, latitude}        |
	// | 22.864359625s | 4313188696 Allocs | map[id][]float64{longitude, latitude}        |
	// |---------------|-------------------|----------------------------------------------|
	// | 13.340439541s | 2416154408 Allocs | map[id][2]float64{longitude, latitude}       |
	// | 13.403035416s | 2413071824 Allocs | map[id][2]float64{longitude, latitude}       |
	// | 13.387366542s | 2436762280 Allocs | map[id][2]float64{longitude, latitude}       |
	// |---------------|-------------------|----------------------------------------------|

	err = e.writeID(id)
	if err != nil {
		err = fmt.Errorf("writeNode().error: the writeID() function returned an error: %v", err)
		return
	}

	err = e.writeCoordinate(longitude)
	if err != nil {
		err = fmt.Errorf("writeNode().error: the writeCoordinate(longitude) function returned an error: %v", err)
		return
	}

	err = e.writeCoordinate(latitude)
	if err != nil {
		err = fmt.Errorf("writeNode().error: the writeCoordinate(latitude) function returned an error: %v", err)
		return
	}

	e.totalOfNodesInTmpFile++

	return
}

// FindNodeByID
//
// English:
//
// Search for longitude and latitude in the temporary file.
//
//	Input:
//	  id: ID of the node sought.
//
//	Output:
//	  longitude: value between ±180 width 7 decimal places;
//	  latitude: value between ±90 width 7 decimal places;
//	  err: pattern object, with io.EOF error when value not found in file
//
// Português:
//
// Procura por longitude e latitude no arquivo temporário.
//
//	Entrada:
//	  id: ID do node procurado.
//
//	Saída:
//	  longitude: valor entre ±180 com 7 casas decimais;
//	  latitude: valor entre ±90 com 7 casas decimais;
//	  err: objeto de padrão, com erro io.EOF quando o valor não é encontrado no arquivo
func (e *Compress) FindNodeByID(id int64) (longitude, latitude float64, err error) {
	i := sort.Search(len(e.memory), func(i int) bool { return e.memory[i][memorySliceAddrID] >= id })
	if i < len(e.memory) && e.memory[i][memorySliceAddrID] == id {

		// English: The searched ID was found in memory and does not need to go through the file lookup.
		// Português: O ID procurado foi encontrado na memória e não necessita passar pela busca no arquivo.
		fileAddressCalculated := e.memory[i][memorySliceAddrOfAddrIntoFile]
		longitude, err = e.readCoordinate(fileAddressCalculated + nodeIdByteSize)
		if err != nil {
			err = fmt.Errorf("FindNodeByID().error: readCoordinate(%v*(8+4+4)+4) function returned an error: %v", fileAddressCalculated, err)
			return
		}

		latitude, err = e.readCoordinate(fileAddressCalculated + nodeIdByteSize + nodeCoordinateByteSize)
		if err != nil {
			err = fmt.Errorf("FindNodeByID().error: readCoordinate(%v*(8+4+4)+4+4) function returned an error: %v", fileAddressCalculated, err)
			return
		}
		return
	}

	// English: Adjust left and right border for binary search.
	// Português: Ajusta a borda inferior e superior para a busca binária.
	if i > 0 {
		i--
	}

	leftBound := e.memory[i][memorySliceAddrOfAddrIntoFile]
	rightBound := e.memory[i+1][memorySliceAddrOfAddrIntoFile]
	longitude, latitude, err = e.binarySearch(leftBound, rightBound, id)
	if err != nil {
		//err = fmt.Errorf("FindNodeByID().error: binarySearch(%v, %v, %v) function returned an error: %v", leftBound, rightBound, id, err)
		return
	}

	return
}

// memoryKeyToFileAddress
//
// English:
//
// Converts the key from the binary search in memory to the address of the temporary file.
//
// Português:
//
// Converte a chave da busca binária em memória para endereço do arquivo temporário.
//func (e *Compress) memoryKeyToFileAddress(key int) (addr int64) {
//	return int64(key)*e.blockSize*nodeDataByteSize + nodeDataPositionStartAtAddress
//}

// binarySearch
//
// English:
//
// Does the binary search in the temporary file.
//
//	Note:
//	  * Code Based on website https://golangprojectstructure.com/super-speed-up-with-binary-search/
//
// Português:
//
// Faz a busca binária no arquivo temporário.
//
//	Nota:
//	  * Baseado no código do site https://golangprojectstructure.com/super-speed-up-with-binary-search/
func (e *Compress) binarySearch(leftBoundFileAddr, rightBoundFileAddr, nodeIdToFind int64) (longitude, latitude float64, err error) {
	if rightBoundFileAddr >= leftBoundFileAddr {
		// todo: The first attempts to simplify the formula gave an error. Stayed for another day.
		fileAddress := leftBoundFileAddr + (((rightBoundFileAddr-leftBoundFileAddr)/nodeDataByteSize)/2)*nodeDataByteSize
		fileAddressCalculated := fileAddress

		var idFound int64
		idFound, err = e.readID(fileAddressCalculated)
		if err != nil {
			err = fmt.Errorf("binarySearch().error: readID(%v*(8+4+4)) function returned an error: %v", fileAddress, err)
			return
		}

		if idFound == nodeIdToFind {
			longitude, err = e.readCoordinate(fileAddressCalculated + nodeIdByteSize)
			if err != nil {
				err = fmt.Errorf("binarySearch().error: readCoordinate(%v*(8+4+4)+4) function returned an error: %v", fileAddress, err)
				return
			}

			latitude, err = e.readCoordinate(fileAddressCalculated + nodeIdByteSize + nodeCoordinateByteSize)
			if err != nil {
				err = fmt.Errorf("binarySearch().error: readCoordinate(%v*(8+4+4)+4+4) function returned an error: %v", fileAddress, err)
				return
			}

			return
		}

		if idFound > nodeIdToFind {
			return e.binarySearch(leftBoundFileAddr, fileAddress-nodeDataByteSize, nodeIdToFind)
		}

		return e.binarySearch(fileAddress+nodeDataByteSize, rightBoundFileAddr, nodeIdToFind)
	}

	err = io.EOF
	return
}

// MountIndexIntoFile
//
// English:
//
// Salva os índices no arquivo temporário.
// See the explanation on the Init() function for more details.
//
//	Note:
//	  * Indexes are blocks with ranges of IDs to help calculate the address of the ID within the temporary file.
//	  * Indexes are loaded into memory for better performance.
//
// Português:
//
// Salva os índices no arquivo temporário.
// Veja a explicação na função Init() para mais detalhes.
//
//	Nota:
//	  * Índices são blocos com intervalos de IDs para ajudar a calcular o endereço do ID dentro do arquivo temporário.
//	  * Índices são carregados em memória para maior desempenho.
func (e *Compress) MountIndexIntoFile() (err error) {
	e.nodeReadDataPosition = nodeDataPositionStartAtAddress

	//(place * (id space + lon space + lat space)) + (version + total nodes + block size + total indexes + index addr) = addr
	//(place * (8 + 4 + 4)) + (8 + 8 + 8 + 8 + 8) = addr
	//(place * 16) + 40 = addr

	// place = (address - (version + total nodes + block size + total indexes + index addr)) / (id space + lon space + lat space)
	// place = (address - (8 + 8 + 8 + 8 + 8)) / (8 + 4 + 4)
	// place = (address - 40) / 16

	// English: points to the last node inserted.
	// Português: aponta para o último node inserido.
	lastNodeAddr := e.nodeWriteDataPosition - nodeDataByteSize

	// English: e.totalIndexIntoFile-1 the last index does not mathematically correspond to the size of the block, it
	// corresponds to the last ID inserted in the file, or it will be outside the search window.
	// Português: e.totalIndexIntoFile-1 o último índice não corresponde matematicamente ao tamanho do bloco, corresponde
	// ao último ID inserido no arquivo, ou o mesmo ficará fora da janela de busca.
	for i := int64(0); i != e.totalIndexIntoFile-1; i += 1 {
		e.nodeReadDataPosition = i*e.blockSize*nodeDataByteSize + nodeDataPositionStartAtAddress
		_, err = e.file.ReadAt(e.dataFile, e.nodeReadDataPosition)
		if err != nil {
			return
		}

		_, err = e.file.WriteAt(e.dataFile, e.nodeWriteDataPosition)
		if err != nil {
			return
		}
		e.nodeWriteDataPosition += nodeIdByteSize

		// Escreve o endereço do ID no arquivo
		binary.LittleEndian.PutUint64(e.dataFile, uint64(e.nodeReadDataPosition))
		_, err = e.file.WriteAt(e.dataFile, e.nodeWriteDataPosition)
		if err != nil {
			return
		}
		e.nodeWriteDataPosition += nodeIdByteSize
	}

	// English: Write the last node ID in the index, or the last data after the block ID will not be found in the search
	// Português: Escreve o ID do ultimo node no índice, ou os últimos dados depois do ID do bloco não serão encontrados
	// na busca
	_, err = e.file.ReadAt(e.dataFile, lastNodeAddr)
	if err != nil {
		return
	}

	_, err = e.file.WriteAt(e.dataFile, e.nodeWriteDataPosition)
	if err != nil {
		return
	}
	e.nodeWriteDataPosition += nodeIdByteSize

	// Escreve o endereço do ID no arquivo
	binary.LittleEndian.PutUint64(e.dataFile, uint64(lastNodeAddr))
	_, err = e.file.WriteAt(e.dataFile, e.nodeWriteDataPosition)
	if err != nil {
		return
	}
	e.nodeWriteDataPosition += nodeIdByteSize

	return
}

// IndexToMemory
//
// English:
//
// Loads the indexes contained in the temporary file into memory.
//
// Português:
//
// Carrega os índices contidos no arquivo temporário na memória.
func (e *Compress) IndexToMemory() (err error) {
	var id, addr int64
	for i := int64(0); i != e.totalIndexIntoFile; i++ {

		// Lê o ID do dado
		_, err = e.file.ReadAt(e.dataFile, e.nodeWriteDataPosition)
		if err != nil && err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		id = int64(binary.LittleEndian.Uint64(e.dataFile))

		e.nodeWriteDataPosition += nodeIdByteSize

		// Lê o endereço do dado
		_, err = e.file.ReadAt(e.dataFile, e.nodeWriteDataPosition)
		if err != nil {
			return
		}

		e.nodeWriteDataPosition += nodeIdByteSize

		addr = int64(binary.LittleEndian.Uint64(e.dataFile))
		e.memory = append(
			e.memory,
			[2]int64{
				memorySliceAddrID:             id,
				memorySliceAddrOfAddrIntoFile: addr,
			},
		)
	}

	return
}

// writeID
//
// English:
//
// Writes the node ID in binary format in the temporary file.
//
// Português:
//
// Escreve o ID do node em formato binário no arquivo temporário.
func (e *Compress) writeID(id int64) (err error) {
	binary.LittleEndian.PutUint64(e.dataFile, uint64(id))

	// ID do way contido no node
	_, err = e.file.WriteAt(e.dataFile, e.nodeWriteDataPosition)

	e.nodeWriteDataPosition += int64ByteSize
	return
}

// readID
//
// English:
//
// Reads node ID from temp file.
//
// Português:
//
// Lê o ID do node no arquivo temporário.
func (e *Compress) readID(nodeReadDataPosition int64) (id int64, err error) {
	_, err = e.file.ReadAt(e.dataFile, nodeReadDataPosition)
	if err != nil {
		return
	}

	id = int64(binary.LittleEndian.Uint64(e.dataFile))
	return
}

// writeCoordinate
//
// English:
//
// Write the coordinate to the temporary file.
//
// Português:
//
// Escreve a coordenada no arquivo temporário.
func (e *Compress) writeCoordinate(coordinate float64) (err error) {
	negativeNumber := coordinate < 0
	if negativeNumber {
		coordinate *= -1.0
	}

	binary.LittleEndian.PutUint64(e.dataFile, uint64(coordinate*decimalPlaces))
	if negativeNumber {
		e.dataFile[mostSignificantByte] = e.dataFile[mostSignificantByte] | mostSignificantBit
	}

	_, err = e.file.WriteAt(e.dataFile[:4], e.nodeWriteDataPosition)
	if err != nil {
		return
	}
	e.nodeWriteDataPosition += nodeCoordinateByteSize
	return
}

// readCoordinate
//
// English:
//
// Reads the coordinate from the temporary file.
//
// Português:
//
// Lê a coordenada no arquivo temporário.
func (e *Compress) readCoordinate(nodeReadDataPosition int64) (coordinate float64, err error) {
	_, err = e.file.ReadAt(e.dataCoordinate, nodeReadDataPosition)
	if err != nil {
		return
	}

	copy(e.dataFile[0:nodeCoordinateByteSize], e.dataCoordinate[0:nodeCoordinateByteSize])
	e.dataFile[nodeCoordinateByteSize+0] = 0x00
	e.dataFile[nodeCoordinateByteSize+1] = 0x00
	e.dataFile[nodeCoordinateByteSize+2] = 0x00
	e.dataFile[nodeCoordinateByteSize+3] = 0x00

	negativeNumber := e.dataFile[mostSignificantByte]&mostSignificantBit == mostSignificantBit
	e.dataFile[mostSignificantByte] = e.dataFile[mostSignificantByte] & mostSignificantBitTwoComplements

	coordinate = float64(int64(binary.LittleEndian.Uint64(e.dataFile))) / decimalPlaces
	if negativeNumber {
		coordinate *= -1
	}

	return
}

// WriteFileHeaders
//
// English:
//
// Write configuration data at the beginning of the file.
//
// Português:
//
// Escreve os dados de configuração no início do arquivo.
func (e *Compress) WriteFileHeaders() (err error) {
	err = e.writeHeaderVersion()
	if err != nil {
		err = fmt.Errorf("WriteFileHeaders().error: the writeHeaderVersion() function returned an error: %v", err)
		return
	}

	err = e.writeHeaderTotalNodes()
	if err != nil {
		err = fmt.Errorf("WriteFileHeaders().error: the writeHeaderTotalNodes() function returned an error: %v", err)
		return
	}

	err = e.writeHeaderBlockSize()
	if err != nil {
		err = fmt.Errorf("WriteFileHeaders().error: the writeHeaderBlockSize() function returned an error: %v", err)
		return
	}

	err = e.writeHeaderTotalIndexIntoFile()
	if err != nil {
		err = fmt.Errorf("WriteFileHeaders().error: the writeHeaderTotalIndexIntoFile() function returned an error: %v", err)
		return
	}

	err = e.writeHeaderIndexesAddress()
	if err != nil {
		err = fmt.Errorf("WriteFileHeaders().error: the writeHeaderIndexesAddress() function returned an error: %v", err)
		return
	}

	return
}

// ReadFileHeaders
//
// English:
//
// Read the configuration data at the beginning of the file.
//
// Português:
//
// Lê os dados de configuração no início do arquivo.
func (e *Compress) ReadFileHeaders() (err error) {
	err = e.readHeaderVersion()
	if err != nil {
		err = fmt.Errorf("ReadFileHeaders().error: the readHeaderVersion() function returned an error: %v", err)
		return
	}

	err = e.readHeaderTotalNodes()
	if err != nil {
		err = fmt.Errorf("ReadFileHeaders().error: the readHeaderTotalNodes() function returned an error: %v", err)
		return
	}

	err = e.readHeaderBlockSize()
	if err != nil {
		err = fmt.Errorf("ReadFileHeaders().error: the readHeaderBlockSize() function returned an error: %v", err)
		return
	}

	err = e.readHeaderTotalIndexIntoFile()
	if err != nil {
		err = fmt.Errorf("ReadFileHeaders().error: the readHeaderTotalIndexIntoFile() function returned an error: %v", err)
		return
	}

	err = e.readHeaderIndexesAddress()
	if err != nil {
		err = fmt.Errorf("ReadFileHeaders().error: the readHeaderIndexesAddress() function returned an error: %v", err)
		return
	}

	return
}

// testVersionStringSize
//
// English:
//
// Test file version compatibility.
//
// Português:
//
// Testa a compatibilidade da versão do arquivo.
func (e *Compress) testVersionStringSize() (err error) {
	if len(headerVersion) != headerVersionByteSize {
		err = errors.New("constant header version is always 8 bytes long")
		return
	}

	return
}

// writeHeaderVersion
//
// English:
//
// # Write code version in config header
//
// Português:
//
// Escreve a versão do código no cabeçalho de configuração
func (e *Compress) writeHeaderVersion() (err error) {
	err = e.testVersionStringSize()
	if err != nil {
		return
	}

	_, err = e.file.WriteAt([]byte(headerVersion), headerVersionAddress)
	return
}

// readHeaderVersion
//
// English:
//
// # Read code version in config header
//
// Português:
//
// Lê a versão do código no cabeçalho de configuração
func (e *Compress) readHeaderVersion() (err error) {
	_, err = e.file.ReadAt(e.dataFile, headerVersionAddress)
	if err != nil {
		return
	}

	if string(e.dataFile) != headerVersion {
		err = fmt.Errorf("file version header does not match code version: %v != %v", string(e.dataFile), headerVersion)
		return
	}

	return
}

// writeHeaderTotalNodes
//
// English:
//
// # Write the total number of nodes in the configuration header
//
// Português:
//
// Escreve a quantidade total de nodes no cabeçalho de configuração
func (e *Compress) writeHeaderTotalNodes() (err error) {
	binary.LittleEndian.PutUint64(e.dataFile, uint64(e.totalOfNodesInTmpFile))
	_, err = e.file.WriteAt(e.dataFile, headerTotalNodesAddress)
	return
}

// readHeaderTotalNodes
//
// English:
//
// # Read the total number of nodes in the configuration header
//
// Português:
//
// Lê a quantidade total de nodes no cabeçalho de configuração
func (e *Compress) readHeaderTotalNodes() (err error) {
	_, err = e.file.ReadAt(e.dataFile, headerTotalNodesAddress)
	if err != nil {
		return
	}

	e.totalOfNodesInTmpFile = int64(binary.LittleEndian.Uint64(e.dataFile))
	return
}

// writeHeaderBlockSize
//
// English:
//
// # Writes block size of indexes in config header
//
// Português:
//
// Escreve o tamanho do bloco de índices no cabeçalho de configuração
func (e *Compress) writeHeaderBlockSize() (err error) {
	binary.LittleEndian.PutUint64(e.dataFile, uint64(e.blockSize))
	_, err = e.file.WriteAt(e.dataFile, headerBlockSizeAddress)
	return
}

// readHeaderBlockSize
//
// English:
//
// # Reads block size of indices from config header
//
// Português:
//
// Lê o tamanho do bloco de índices no cabeçalho de configuração
func (e *Compress) readHeaderBlockSize() (err error) {
	_, err = e.file.ReadAt(e.dataFile, headerBlockSizeAddress)
	if err != nil {
		return
	}

	e.blockSize = int64(binary.LittleEndian.Uint64(e.dataFile))
	return
}

// writeHeaderTotalIndexIntoFile
//
// English:
//
// # Writes total indexes in configuration header
//
// Português:
//
// Escreve o total de índices no cabeçalho de configuração
func (e *Compress) writeHeaderTotalIndexIntoFile() (err error) {
	e.totalIndexIntoFile = e.totalOfNodesInTmpFile / e.blockSize
	if e.totalOfNodesInTmpFile%e.blockSize != 0 {
		e.totalIndexIntoFile += 1
	}
	binary.LittleEndian.PutUint64(e.dataFile, uint64(e.totalIndexIntoFile))
	_, err = e.file.WriteAt(e.dataFile, headerTotalIndexAddress)
	return
}

// readHeaderTotalIndexIntoFile
//
// English:
//
// # Read the total indexes in configuration header
//
// Português:
//
// Lê o total de índices no cabeçalho de configuração
func (e *Compress) readHeaderTotalIndexIntoFile() (err error) {
	_, err = e.file.ReadAt(e.dataFile, headerTotalIndexAddress)
	if err != nil {
		return
	}

	e.totalIndexIntoFile = int64(binary.LittleEndian.Uint64(e.dataFile))
	return
}

// writeHeaderIndexesAddress
//
// English:
//
// # Write the address of the indexes in the configuration header
//
// Português:
//
// Escreve o endereço dos índices no cabeçalho de configuração
func (e *Compress) writeHeaderIndexesAddress() (err error) {
	binary.LittleEndian.PutUint64(e.dataFile, uint64(e.nodeWriteDataPosition))
	_, err = e.file.WriteAt(e.dataFile, headerIndexesPositionAddress)
	return
}

// readHeaderIndexesAddress
//
// English:
//
// # Reads address of indexes in the configuration header
//
// Português:
//
// Lê o endereço dos índices no cabeçalho de configuração
func (e *Compress) readHeaderIndexesAddress() (err error) {
	_, err = e.file.ReadAt(e.dataFile, headerIndexesPositionAddress)
	if err != nil {
		return
	}

	e.nodeWriteDataPosition = int64(binary.LittleEndian.Uint64(e.dataFile))
	return
}
