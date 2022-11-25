package downloadApiV06

import (
	"encoding/xml"
	"fmt"
	"goosm/goosm"
	"io"
	"net/http"
	"strconv"
)

// osmNode
//
// English:
//
// type generated with the help of the website https://www.onlinetool.io/xmltogo/ and data extracted from the URL
// https://www.openstreetmap.org/api/0.6/node/273316
//
// Português:
//
// tipo gerado com ajuda do site https://www.onlinetool.io/xmltogo/ e dado extraído da URL
// https://www.openstreetmap.org/api/0.6/node/273316
type osmNode struct {
	XMLName xml.Name `xml:"osm"`
	Node    struct {
		ID      string `xml:"id,attr"`
		Visible string `xml:"visible,attr"`
		Lat     string `xml:"lat,attr"`
		Lon     string `xml:"lon,attr"`
		Tag     []struct {
			K string `xml:"k,attr"`
			V string `xml:"v,attr"`
		} `xml:"tag"`
	} `xml:"node"`
}

// osmWay
//
// English:
//
// type generated with the help of the website https://www.onlinetool.io/xmltogo/ and data extracted from the URL
// https://www.openstreetmap.org/api/0.6/way/1099830665
//
// Português:
//
// tipo gerado com ajuda do site https://www.onlinetool.io/xmltogo/ e dado extraído da URL
// https://www.openstreetmap.org/api/0.6/way/1099830665
type osmWay struct {
	XMLName xml.Name `xml:"osm"`
	Way     struct {
		ID         string `xml:"id,attr"`
		Visible    string `xml:"visible,attr"`
		NodeIdList []struct {
			Ref string `xml:"ref,attr"`
		} `xml:"nd"`
		Tag []struct {
			K string `xml:"k,attr"`
			V string `xml:"v,attr"`
		} `xml:"tag"`
	} `xml:"way"`
}

// osmRelation
//
// English:
//
// type generated with the help of the website https://www.onlinetool.io/xmltogo/ and data extracted from the URL
// https://www.openstreetmap.org/api/0.6/relation/5577786
//
// Português:
//
// tipo gerado com ajuda do site https://www.onlinetool.io/xmltogo/ e dado extraído da URL
// https://www.openstreetmap.org/api/0.6/relation/5577786
type osmRelation struct {
	XMLName  xml.Name `xml:"osm"`
	Relation struct {
		ID      string `xml:"id,attr"`
		Visible string `xml:"visible,attr"`
		Member  []struct {
			Type string `xml:"type,attr"`
			Ref  string `xml:"ref,attr"`
			Role string `xml:"role,attr"`
		} `xml:"member"`
		Tag []struct {
			K string `xml:"k,attr"`
			V string `xml:"v,attr"`
		} `xml:"tag"`
	} `xml:"relation"`
}

// DownloadApiV06
//
// English:
//
// # Object compatible with InterfaceDownloadOsm interface
//
// Português:
//
// Objeto compatível com a interface InterfaceDownloadOsm
type DownloadApiV06 struct{}

// DownloadNode
//
// English:
//
// # Download the initialized node for use
//
// Português:
//
// Faz o download do node inicializado para uso
func (e DownloadApiV06) DownloadNode(id int64) (node goosm.Node, err error) {
	//id: 273316
	var data []byte
	data, err = e.download(id, "node")
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadNode().download(%v, %v).error: %v", id, "node", err)
		return
	}

	var nodeXml osmNode
	err = xml.Unmarshal(data, &nodeXml)
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadNode().Unmarshal().error: %v", err)
		return
	}

	var longitude, latitude float64
	longitude, err = strconv.ParseFloat(nodeXml.Node.Lon, 64)
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadNode().ParseFloat(0).error: %v", err)
		return
	}

	latitude, err = strconv.ParseFloat(nodeXml.Node.Lat, 64)
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadNode().ParseFloat(1).error: %v", err)
		return
	}

	tags := make(map[string]string)
	for _, tag := range nodeXml.Node.Tag {
		tags[tag.K] = tag.V
	}
	node.Init(id, longitude, latitude, &tags)

	return
}

// DownloadWay
//
// English:
//
// Downloads the initialized way for use.
//
// Português:
//
// Faz o download do way inicializado para uso.
func (e DownloadApiV06) DownloadWay(id int64) (way goosm.Way, err error) {
	//id: 1099830665

	var node goosm.Node
	var nodeID int64
	var data []byte
	data, err = e.download(id, "way")
	err = fmt.Errorf("downloadApiV06.DownloadWay().download(%v, %v).error: %v", id, "way", err)
	if err != nil {
		return
	}

	var wayXml osmWay
	err = xml.Unmarshal(data, &wayXml)
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadWay().Unmarshal().error: %v", err)
		return
	}

	way.Loc = make([][2]float64, len(wayXml.Way.NodeIdList))
	for nodeKey, nodeRef := range wayXml.Way.NodeIdList {
		nodeID, err = strconv.ParseInt(nodeRef.Ref, 10, 64)
		if err != nil {
			err = fmt.Errorf("downloadApiV06.DownloadWay().ParseInt().error: %v", err)
			return
		}

		node, err = e.DownloadNode(nodeID)
		if err != nil {
			err = fmt.Errorf("downloadApiV06.DownloadWay().DownloadNode(%v).error: %v", nodeID, err)
			return
		}

		way.Loc[nodeKey] = [2]float64{node.Loc[0], node.Loc[1]}
	}

	tags := make(map[string]string)
	for _, tag := range wayXml.Way.Tag {
		tags[tag.K] = tag.V
	}
	way.Tag = tags
	err = way.Init()
	if err != nil {
		err = fmt.Errorf("downloadApiV06.DownloadWay().way.Init().error: %v", err)
		return
	}

	return
}

// DownloadRelation
//
// English:
//
// Downloads the initialized relation for use.
//
// Português:
//
// Faz o download da relation inicializado para uso.
func (e DownloadApiV06) DownloadRelation(id int64) (relation goosm.Relation, err error) {
	//id: 5577786
	var data []byte
	data, err = e.download(id, "relation")
	if err != nil {
		return
	}

	var relationXml osmRelation
	err = xml.Unmarshal(data, &relationXml)
	if err != nil {
		return
	}

	//todo: fazer

	return
}

// download
//
// English:
//
// # Download the element in byte form
//
// Português:
//
// Faz o download do elemento na forma de byte
func (e DownloadApiV06) download(id int64, elementType string) (data []byte, err error) {
	var resp *http.Response
	resp, err = http.Get("https://www.openstreetmap.org/api/0.6/" + elementType + "/" + strconv.FormatInt(id, 10))
	if err != nil {
		return
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	_ = resp.Body.Close()
	return
}
