package goosm

import (
	"math"
	"time"
)

const (
	Longitude            = 0
	Latitude             = 1
	DB_API_CONFIG string = "apiConfig"

	// Address from server test
	DB_SERVER_TEST string = "127.0.0.1"

	// Database into server test
	// Warning, this database is created and deleted in every test
	DB_DATABASE_TEST string = "gOsm_test"

	DB_CONFIGURATION_COLLECTIONS string = "configuration"

	DB_USER_LOGIN_COLLECTIONS              string = "users"
	DB_USER_LOGIN_SECURITY_LOG_COLLECTIONS string = "usersLog"

	// It is necessary to wait a while between the bank's destruction and the creation of a new
	DELAY_TIME_BETWEEN_TESTS time.Duration = 1

	// Collection nodes from importation of xml
	DB_OSM_FILE_NODES_COLLECTIONS string = "nodes"

	DB_OSM_FILE_PSEUDO_ID_COLLECTIONS     string = "pseudoIds"
	DB_OSM_FILE_PSEUDO_ID_COLLECTIONS_LOG string = "pseudoIdsLog"

	DB_OSM_ROCKSDB_SUBDIR            int    = 8
	DB_OSM_ROCKSDB_COLLECTIONS       string = "./rocksdb"
	DB_OSM_ROCKSDB_NODES_COLLECTIONS string = "/nodes"

	//todo mudar isto
	DB_OSM_FILE_POLYGONS_COLLECTIONS                    string = "polygons"
	DB_OSM_FILE_POLYGONS_CONCAVE_HULL_COLLECTIONS       string = "polygonsConcaveHull"
	DB_OSM_FILE_POLYGONS_CONVEX_HULL_COLLECTIONS        string = "polygonsConvexHull"
	DB_OSM_FILE_POLYGONS_SURROUNDINGS_COLLECTIONS       string = "polygonsSurroundings"
	DB_OSM_FILE_POLYGONS_SURROUNDINGS_LEFT_COLLECTIONS  string = "polygonsSurroundingsLeft"
	DB_OSM_FILE_POLYGONS_SURROUNDINGS_RIGHT_COLLECTIONS string = "polygonsSurroundingsRight"

	DB_OSM_TEMPORARY_POINTS_COLLECTIONS string = "tmpPoint"
	DB_OSM_ERROR_LOG_COLLECTIONS        string = "errorLog"
	DB_OSM_IMPORT_LOG_COLLECTIONS       string = "importLog"
	DB_OSM_SOURCE_FILES_LOG_COLLECTIONS string = "sourceFile"

	// Collection of names of tags from nodes and ways
	DB_OSM_FILE_TAG_NAME_COLLECTIONS string = "tagNames"

	DB_IBGE_FILE_COUNTY_COLLECTIONS       string = "ibgeCounty"
	DB_IBGE_FILE_DISTRICT_COLLECTIONS     string = "ibgeDistrict"
	DB_IBGE_FILE_NEIGHBORHOOD_COLLECTIONS string = "ibgeNeighborhood"
	DB_IBGE_FILE_COUNTY_DATA_COLLECTIONS  string = "ibgeCountyAllData"

	// Collection ways from importation of xml
	DB_OSM_FILE_WAYS_COLLECTIONS            string = "ways"
	DB_OSM_FILE_RELATIONS_COLLECTIONS       string = "relations"
	DB_OSM_FILE_MULTIPOLYGONS_COLLECTIONS   string = "multiPolygons"
	DB_OSM_FILE_POLYGONS_TMP_COLLECTIONS    string = "polygonsTmp"
	DB_OSM_FILE_RELATIONS_COLLECTIONS_ERROR string = "relationsError"

	DB_OSM_FILE_RAM_COLLECTIONS = "ramVar"

	DB_TEST_POINTS_COLLECTIONS string = "points"

	DB_TEST_GEOJSON_POLYGONS_COLLECTIONS      string = "geoJsonMultiPolygons"
	DB_TEST_GEOJSON_FEATURES_COLLECTIONS      string = "geoJsonFeatures"
	DB_TEST_GEOJSON_POLYGONS_TAGS_COLLECTIONS string = "geoJsonMultiPolygonsTags"

	DB_GEOJSON_POLYGONS_COLLECTIONS              string = "geoJsonMultiPolygons"
	DB_GEOJSON_FEATURES_COLLECTIONS              string = "geoJsonFeatures"
	DB_GEOJSON_POLYGONS_TAGS_COLLECTIONS         string = "geoJsonMultiPolygonsTags"
	DB_GEOJSON_CONCAVE_HULL_POLYGONS_COLLECTIONS string = "geoJsonConcaveHullPolygon"
	DB_GEOJSON_CONVEX_HULL_POLYGONS_COLLECTIONS  string = "geoJsonConvexHullPolygon"

	DB_TEST_GEOJSON_CONCAVE_HULL_POLYGONS_COLLECTIONS string = "geoJsonConcaveHullPolygon"
	DB_TEST_GEOJSON_CONCAVE_HULL_FEATURES_COLLECTIONS string = "geoJsonConcaveHullFeatures"

	DB_TEST_CONCAVE_HULL_POLYGONS_COLLECTIONS string = "concaveHullMultiPolygons"
	DB_TEST_CONCAVE_HULL_FEATURES_COLLECTIONS string = "concaveHullFeatures"

	// Degrees symbol for human notation
	// Warning, changing this value affect an innumerable amount of testing
	DEGREES string = "°"

	// Minutes from degrees symbol for human notation
	// Warning, changing this value affect an innumerable amount of testing
	MINUTES string = "´"

	// Seconds from degrees symbol for human notation
	// Warning, changing this value affect an innumerable amount of testing
	SECONDS string = "´´"

	// Rads symbol from human notation
	// Warning, changing this value affect an innumerable amount of testing
	RADIANS string = "rad"

	// Semi-axes of WGS-84 geoidal reference
	// The default datum MongoDB uses to calculate geometry over an Earth-like sphere. MongoDB uses the WGS84 datum for
	// geospatial queries on GeoJSON objects. See the “EPSG:4326: WGS 84” specification:
	// http://spatialreference.org/ref/epsg/4326/.
	// Warning, changing this value affect an innumerable amount of testing and MongoDB results.
	GEOIDAL_MAJOR float64 = WGS84_a
	GEOIDAL_MINOR float64 = WGS84_b

	// Geoide WGS84, major semiaxis in meters
	// Warning, changing this value affect an innumerable amount of testing
	WGS84_a float64 = 6378137.0

	// Geoide WGS84, minor semiaxis in meters
	// Warning, changing this value affect an innumerable amount of testing
	WGS84_b float64 = 6356752.314245

	// Geoide GRS_80, major semiaxis in meters
	// Warning, changing this value affect an innumerable amount of testing
	GRS_80_a float64 = 6378137.0

	// Geoide GRS_80, major semiaxis in meters
	// Warning, changing this value affect an innumerable amount of testing
	GRS_80_b float64 = 6356752.314140

	// Minimal latitude from point
	// Warning, you should not touch
	MIN_LAT float64 = math.Pi * -80.0 / 180.0 //math.Pi / 2 * -1

	// Maximal latitude from point
	// Warning, you should not touch
	MAX_LAT float64 = math.Pi * 84.0 / 180.0 //math.Pi / 2

	// Minimal longitude from point
	// Warning, you should not touch
	MIN_LON float64 = math.Pi * -1

	// Maximal longitude from point
	// Warning, you should not touch
	MAX_LON float64 = math.Pi

	// Warning, you should not touch! Never!
	FIND_ALL_QUERY = 0
	FIND_ALL_LIMIT = 1
	FIND_ALL_SKIP  = 2
	FIND_ALL_SORT  = 3

	DB_GOSM_GEOJSON_FEATURES_COLLECTIONS string = "geoJsonGOsm"
)
