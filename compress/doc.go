// Package compress
//
// # English:
//
//	This package stores geographic coordinates used in building OpenStreetMap in a binary file designed to be
//	efficient in processing ways, in addition it providing an interface called for each processed way.
//
// The problem:
//
//	OpenStreetMap serves a binary file containing all the nodes, 7.9 trillion the last time I downloaded the complete
//	world, identified by an ID and a coordinate, and then presents all the ways, containing only the node ID, which
//	generates a search by NODE_ID for each node contained in the way, which slows down processing.
//
//	This package creates a binary search to optimize map processing and return a ready-to-use way.
//
// Geo coordinate compression for maps:
//
//	   On a planned map, the largest number, in number of decimal places, present on the map is +/-180˚ in longitude and
//	   +/-90˚ in latitude, with 7 decimal places.
//
//		  For data compression, the floating point number of the coordinate is multiplied by 10,000,000 and then converted
//		  to an integer, losing the decimal part, so the largest number saved is the integer +/-1,800,000,000.
//		  With this, it can be represented by the group of four bytes X110 1011 0100 1001 1101 0010 0000 0000, where the
//		  most significant bit, `X` is never used, and can be used to indicate positive or negative sign, that is, X= 1
//		  represents a negative number and X=0 a positive number (the rule of two was not used).
//
// File format:
//
//		Header: 24 bytes
//	   version: 8 bytes
//		  total of nodes in a file: 8 bytes
//		  total block size: 8 bytes
//	   total index into file 8 bytes
//		  start index address: 8 bytes
//
//		Data block:
//		  node.ID: 8 bytes
//		  node.Longitude: 4 bytes
//		  node.Latitude: 4 bytes
//
//		Index block:
//		  Indexes are a fixed-size block used for in-memory indexing, where a block represents the address of the node in
//		  the file.
//		  For example:
//		  For the list of node ID 1 to 100 in the file, and block size 10, memory will contain the values 1, 11, 21, ...,
//		  81, 91.
//		  Compress.FindNodeByID(75), it will first do a `secondary binary search, in memory` search and find indices 7 and 8
//		  for the left edge and right edge.
//		  Applying the formula: (address * 16) + header size, 40
//		  the left edge will have ID 71, (memory[7][0]), and address 1160, (memory[7][1]),
//		  the right edge will have ID 81, (memory[8][0]) and address (memory[8][1]) 1320
//		  Therefore, the ID sought will be between addresses 1160 and 1320 of the binary file on disk.
//		  On disk, each address is 8 bytes for ID + 4 bytes for longitude + 4 bytes for latitude.
//
// # Português:
//
//	Este pacote arquiva coordenadas geográficas usadas na construção do OpenStreetMap em um arquivo binário feito para
//	ser eficiente no processamento de ways, além de fornecer uma interface chamada a cada way processado.
//
// O problema:
//
//	O OpenStreetMap entrega um arquivo binário contendo todos os nodes, 7.9 trilhões a última vez que eu baixei o
//	mundo completo, identificado por um ID e uma coordenada, e em seguida, apresenta todos os ways, contendo apenas
//	o ID do node, oq ue gera uma busca por NODE_ID para cada node contido no way, o que deixa o processamento lento.
//
//	Este pacote cria uma busca binária em memória e em arquivo para otimizar o processamento do mapa e devolver um way
//	montado e pronto para uso.
//
// Compactação de coordenada geográfica para mapas:
//
//	   Em um mapa planificado, o maior número, em quantidade de casas decimais, presente no mapa é +/-180˚ na longitude
//	   e +/-90˚ na latitude, com 7 casas decimais.
//
//		  Para a compactação de dados, o número de ponto flutuante da coordenada é multiplicado por 10.000.000 e em seguida
//		  é convertido em inteiro, perdendo a parte decimal, logo, o maior número salvo é o inteiro +/-1.800.000.000.
//		  Com isto, pode ser representado pelo grupo de quatro bytes X110 1011 0100 1001 1101 0010 0000 0000, onde o bit
//		  mais significativo, `X` nunca é usado, e pode ser usado para indicar sinal de positivo ou negativo, ou seja, X=1
//		  representa um número negativo e X=0 um número positivo (não foi usada a regra de dois).
//
// Formato do arquivo:
//
//		Header: 24 bytes
//	   version: 8 bytes
//		  total of nodes in a file: 8 bytes
//		  total block size: 8 bytes
//	   total index into file 8 bytes
//		  start index address: 8 bytes
//
//		Data block:
//		  node.ID: 8 bytes
//		  node.Longitude: 4 bytes
//		  node.Latitude: 4 bytes
//
//		Index block:
//		  Índices são um bloco de tamanho fixo, usado para uma indexação em memória, onde um bloco representa o endereço do
//		  node no arquivo.
//		  Por exemplo:
//		  Para a lista de node ID 1 a 100, no arquivo, e tamanho do bloco 10, memory conterá os valores 1, 11, 21, ...,
//		  81, 91.
//		  Compress.FindNodeByID(75), primeiro fará uma busca em `secondary binary search, in memory` e encontrará os índices
//		  7 e 8 para a borda esquerda e a borda direita.
//		  Aplicando a fórmula: (endereço * 16) + tamanho do cabeçalho, 40
//		  a borda esquerda terá o ID 71, (memory[7][0]), e o endereço 1.160, (memory[7][1]),
//		  a borda direita terá o ID 81, (memory[8][0]) e o endereço (memory[8][1]) 1.320
//		  Logo, o ID procurado estará entre os endereços 1.160 e 1.320 do arquivo binário em disco.
//		  No disco, cada endereço tem 8 bytes para ID + 4 bytes para a longitude + 4 bytes para a latitude.
//
// # Drawing:
//
//	Drawing the binary file for better understanding:
//	Desenhando o arquivo binário para melhor entendimento:
//
//	   addr:00000 version 8bytes         --+
//	   addr:00008 total of nodes           |
//	   addr:00016 block size               +- header, configuration
//	   addr:00024 total of index into file |
//	   addr:00032 start index addr       --+
//	   ID:0001 Addr:00040                --+
//	   ID:0002 Addr:00056                  |
//	   ...                                 |
//	   ...                                 |
//	   ID:0070 Addr:01144                  |
//	   ID:0071 Addr:01160                  |
//	   ID:0072 Addr:01176                  |
//	   ID:0073 Addr:01192                  |
//	   ID:0074 Addr:01208                  |
//	   ID:0075 Addr:01224                  +- primary binary search, on disk
//	   ID:0076 Addr:01240                  |
//	   ID:0077 Addr:01256                  |
//	   ID:0078 Addr:01272                  |
//	   ID:0079 Addr:01288                  |
//	   ID:0080 Addr:01304                  |
//	   ID:0081 Addr:01320                  |
//	   ID:0082 Addr:01336                  |
//	   ...                                 |
//	   ...                                 |
//	   ID:0100                           --+
//	   ID:0001:Addr:00040                --+
//	   ID:0011:Addr:00200                  |
//	   ID:0021:Addr:00360                  |
//	   ID:0031:Addr:00520                  |
//	   ID:0041:Addr:00680                  +- secondary binary search, in memory
//	   ID:0051:Addr:00840                  |
//	   ID:0061:Addr:01000                  |
//	   ID:0071:Addr:01160                  |
//	   ID:0081:Addr:01320                  |
//	   ID:0091:Addr:01480                --+
package compress
