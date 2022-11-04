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
	dockerContainer.SetWaitStringWithTimeout(`"msg":"Waiting for connections","attr":{"port":27017`, 20*time.Second)

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
