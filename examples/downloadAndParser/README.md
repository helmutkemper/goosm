# download and parser

## English:

This example downloads a map from the geo frabrik site, sets up a binary search with all nodes and then prepares the 
ways to be used.

In parallel, add all nodes that are not just for building ways in the node collection. 

#### Requirements:

[MongoDB](https://www.mongodb.com/docs/manual/installation/) installed on port 27016 with the `osm` bank free to use.

#### How to run the example:

```shell
  make build
```

## Português:

### Problema

O Open Street Maps têm mapas com um volume grande de dados, o mapa global tem algo em torno de 7.9 trilhões de pontos,
tornando o processamento de dados muito demorado, elevando o custo de manutenção do projeto.

Os dados são arquivados em formato relacional, onde a primeira parte do arquivo contém todos os nodes e a segunda parte 
do arquivo contém todos os ways com os IDs dos nodes contidos na primeira parte do arquivo.

Quando se tenta inserir todos os ~7.9 trilhões de nodes no banco de dados, o tempo de inserção sobe muito a medida que
banco de dados é preenchido.

Para entender o problema, pense em IDs linear, de 1 a 7.9 trilhões. 
No teste, inserir o ID 1 levou µS, mas, a medida que o banco de dados foi preenchido, o tempo de inserção chegou a ms.
Procurar por um ID no banco de dados enfrenta o mesmo problema, o primeiro ID inserido em uma busca tipo, findById(1),
retorna a informação em µS, já a busca pelo último ID inserido, findById(1.000.000.000), retorna o valor em ms.
Isto deixa o processamento de dados muito lento.

### Solução

Criar um arquivo para busca binária, onde todos os nodes são arquivados em binário.

O dado é formado por `ID+Longitude+Latitude` ou `[8]bytes+[4]bytes+[4]bytes`.

Os dados são convertidos em binário antes de serem salvos, usando o pacote binário do golang, pois, o teste de benchmark
mostraram os seguintes desempenhos:

| total process | Memory Allocs     | Type of golang data                          |
|---------------|-------------------|----------------------------------------------|
|  3.830812750s | 1511656176 Allocs | []byte using binary.LittleEndian.PutUint64() |
| 13.403035416s | 2413071824 Allocs | map[id][2]float64{longitude, latitude}       |
| 23.104297458s | 4426114960 Allocs | map[id][]float64{longitude, latitude}        |

Porém, uma busca binária simples, com 7.9 trilhões de IDs ainda seria mais demorada do que o necessário, por isto, ao
final do arquivo, são salvos amostras de IDs para uma segunda busca binária em memória, onde a busca retorna dois 
endereços, a borda esquerda e a borda direita de onde o ID procurado se encontra no arquivo binário, assim, a busca é
sempre limitada a um bloco de tamanho fixo, definido na criação do arquivo binário.

```go
// headers:
// ...
// dados:
// ID:1
// ID:2
// ...
// ...
// ID:1024
// ID:1025
// ...
// ...
// ID:x
// Índices:
// ID:0001:Addr:00040 --+
// ID:1025:Addr:18490   +- busca binária em memória
// ID:2049:Addr:32824 --+
```

### Resultado

Para um arquivo pequeno, o processamento foi reduzido para algo em torno de 1h, usando um Mac Book M1.

### Exemplo

Este exemplo faz o download de um mapa do site geo frabrik, monta uma busca binária com todos os nodes e em seguida 
prepara os nodes e ways para serem usados na forma de [GeoJSon](https://geojson.io)

### Requerimentos:

[MongoDB](https://www.mongodb.com/docs/manual/installation/) instalado na porta 27016 com o banco `osm` livre para uso.

> Há um exemplo de como instalar o `MongoDB` de forma simples, com a ajuda do `docker`

### Como usar este exemplo:

```shell
  make build
```
















