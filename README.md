# goosm

Fast open street maps to database

<!--div align="center">
  <a href="https://www.youtube.com/watch?v=tZbQPdCAtC0"><img src="https://img.youtube.com/vi/tZbQPdCAtC0/0.jpg" width="900px" alt="youtube video"></a>
</div-->

## English

This code shows how to decrease the time of importing `Open Street Maps` maps using binary search to optimize the search 
for data.

### List of examples

| Example             | Description                                                                                          |
|---------------------|------------------------------------------------------------------------------------------------------|
| MongoDB Install     | Install the `MongoDB` database on port `27016` and use the `docker/mongodb/data` folder for the data |
| Download and parser | Download a map and insert the contents into the `MongoDB` database                                   |
| Find one by id      | Shows how to use `MongoDB` driver and capture `GeoJson` from geographic information                  |


### Interfaces

| Name                 | Description                                       |
|----------------------|---------------------------------------------------|
| CompressInterface    | Data compression for binary search                |
| InterfaceDownloadOsm | Download data using Open Street Maps API V0.6     |
| InterfaceConnect     | Database connection, used by node and way objects |
| InterfaceDbNode      | Inserting nodes into the database                 |
| InterfaceDbWay       | Inserting ways into the database                  |

## Português

Este código mostra como diminuir o tempo de importação dos mapas do `Open Street Maps` usando busca binária para 
otimizar a procura por dados

### Lista de exemplos

| Exemplo              | Descrição                                                                                             |
|----------------------|-------------------------------------------------------------------------------------------------------|
| MongoDB Install      | Instala o banco de dados `MongoDB` na porta `27016` e usa a pasta `docker/mongodb/data` para os dados |
| Download and parser  | Faz o download de um mapa e insere o conteúdo no banco de dados `MongoDB`                             |
| Find one by id       | Mostra como usar o driver `MongoDB` e capturar o `GeoJson` da informação geográfica                   |

### Interfaces

| Nome                 | Descrição                                                      |
|----------------------|----------------------------------------------------------------|
| CompressInterface    | Compressão de dados para busca binária                         |
| InterfaceDownloadOsm | Faz o download de dados usando a API V0.6 do Opens Street Maps |
| InterfaceConnect     | Conexão do banco de dados, usada pelos objetos node e way      |
| InterfaceDbNode      | Inserção de nodes no banco de dados                            |
| InterfaceDbWay       | Inserção de ways no banco de dados                             |


