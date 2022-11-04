package datasource

import (
	"errors"
	jwtverify "goosm/businessRules/toModule/JWT"
	"goosm/businessRules/toModule/passwordHash"
	"goosm/businessRules/toModule/uID"
	"goosm/goosm"
	"goosm/module/interfaces"
	mongodbosm "goosm/plugin/osm.mongodb"
	"plugin"
)

type RefList struct {
	Osm      goosm.InterfaceDatabase      `json:"-"`
	Password interfaces.InterfacePassword `json:"-"`
	UniqueID interfaces.InterfaceUID      `json:"-"`
	Jwt      interfaces.InterfaceJWT      `json:"-"`
}

func (e *RefList) GetReferenceFromOsm() (datasource goosm.InterfaceDatabase) {
	return e.Osm
}

// Init (Português): Inicializa o datasource escolhido
//
//	name: tyme Name
//	  KSQLite: Inicializa o banco de dados como sendo o SQLite
func (e *RefList) Init(name Name) (err error) {

	//var path string

	err = errors.New("configure the data source first")

	// Inicializa o objeto Password
	e.Password = &passwordHash.Password{}

	// Inicializa o objeto UID
	e.UniqueID = &uID.UID{}

	// Inicializa o gerador/verificador de JWT
	e.Jwt = &jwtverify.JwtVerify{}
	err = e.Jwt.NewAlgorithm([]byte("colocar em constants")) //fixme
	if err != nil {
		return
	}

	e.Osm = &mongodbosm.MongoDbOsm{}
	_, err = e.Osm.New()

	// Inicializa o banco de dados
	//switch name {
	//
	//case KMongoDB:
	//
	//	path, err = util.FileFindInTree("osm.mongodb.so")
	//	if err != nil {
	//		return
	//	}
	//
	//	err = e.installOsmByPlugin(path)
	//	if err != nil {
	//		return
	//	}
	//
	//case KSQLite:
	//	path, err = util.FileFindInTree("osm.sqlite.so")
	//	if err != nil {
	//		return
	//	}
	//
	//	err = e.installOsmByPlugin(path)
	//	if err != nil {
	//		return
	//	}
	//}

	return
}

func (e *RefList) installOsmByPlugin(pluginPlath string) (err error) {
	var ok bool
	var menu *plugin.Plugin
	var menuSymbol plugin.Symbol

	menu, err = plugin.Open(pluginPlath)
	if err != nil {
		return
	}

	menuSymbol, err = menu.Lookup("Osm")
	if err != nil {
		return
	}

	e.Osm, ok = menuSymbol.(goosm.InterfaceDatabase)
	if ok == false {
		err = errors.New("plugin osm conversion into interface osm has an error")
		return
	}

	_, err = e.Osm.New()
	return
}
