package compress

import (
	"os"
	"testing"
)

// TestCompress_node
//
// English:
//
// # Write a node to file and test the read values
//
// Português:
//
// Escreve um node em arquivo e testa os valores lidos
func TestCompress_node(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Remove("./test.node.tmp")
	})

	var err error
	var id int64
	var longitude, latitude float64

	compress := Compress{}
	compress.Init(10)
	err = compress.Create("./test.node.tmp")
	if err != nil {
		t.Logf("open file error: %v", err)
		t.FailNow()
	}

	err = compress.WriteNode(123, -123.1234567, 12.9876543)
	if err != nil {
		t.Logf("write node error: %v", err)
		t.FailNow()
	}

	id, err = compress.readID(nodeDataPositionStartAtAddress)
	if err != nil {
		t.Logf("read id error: %v", err)
		t.FailNow()
	}

	if id != 123 {
		t.Logf("read id error: 123 != %v", id)
		t.FailNow()
	}

	// address from data start + 8 bytes from ID size = 1˚ coordinate address.
	longitude, err = compress.readCoordinate(nodeDataPositionStartAtAddress + nodeIdByteSize)
	if err != nil {
		t.Logf("read longitude error: %v", err)
		t.FailNow()
	}

	if longitude != -123.1234567 {
		t.Logf("read longitude error: -123.1234567 != %v", longitude)
		t.FailNow()
	}

	// address from data start + 8 bytes from ID size + 4 bytes 1˚ coordinate address = 2˚ coordinate address.
	latitude, err = compress.readCoordinate(nodeDataPositionStartAtAddress + nodeIdByteSize + nodeCoordinateByteSize)
	if err != nil {
		t.Logf("read latitude error: %v", err)
		t.FailNow()
	}

	if latitude != 12.9876543 {
		t.Logf("read latitude error: 12.9876543 != %v", longitude)
		t.FailNow()
	}
}

func TestCompress(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Remove("./test.node.tmp")
	})

	var err error
	var testLimit = 100

	compress := Compress{}
	compress.Init(7)
	err = compress.Create("./test.node.tmp")
	if err != nil {
		t.Logf("open file error: %v", err)
		t.FailNow()
	}
	defer compress.Close()

	// Gera a informação de controle
	nodeList := make([]Node, 0)
	for i := int64(0); i != int64(testLimit); i++ {
		nodeList = append(nodeList, Node{ID: i + 1, Lon: compress.Round((float64(i)*0.00001+2.123456)*-1, 6.0), Lat: compress.Round(float64(i)*0.00001+1.98765, 6.0)})
	}

	for i := 0; i != testLimit; i++ {
		compress.totalOfNodesInTmpFile++

		err = compress.WriteNode(nodeList[i].ID, nodeList[i].Lon, nodeList[i].Lat)
		if err != nil {
			t.Logf("write node error: %v", err)
			t.FailNow()
		}
	}

	err = compress.WriteFileHeaders()
	if err != nil {
		t.Logf("write header error: %v", err)
		t.FailNow()
	}

	err = compress.MountIndexIntoFile()
	if err != nil {
		t.Logf("write index error: %v", err)
		t.FailNow()
	}

	// -------------------------------------------------------------------------------------------------------------------
	// fim da escrita do arquivo
	// -------------------------------------------------------------------------------------------------------------------

	compress.Init(7)
	err = compress.Create("./test.node.tmp")
	if err != nil {
		t.Logf("open file error: %v", err)
		t.FailNow()
	}

	err = compress.ReadFileHeaders()
	if err != nil {
		t.Logf("headers error: %v", err)
		t.FailNow()
	}

	err = compress.IndexToMemory()
	if err != nil {
		t.Logf("IndexToMemory() error: %v", err)
		t.FailNow()
	}

	var id int64
	var longitude, latitude float64
	for i := 0; i != testLimit; i++ {
		id = int64(i + 1)
		longitude, latitude, err = compress.FindNodeByID(id)
		if err != nil {
			t.Logf("findID() error: %v", err)
			t.FailNow()
		}

		if compress.Round(longitude, 6.0) != compress.Round((float64(i)*0.00001+2.123456)*-1, 6.0) {
			t.Logf("longitude error: %v != %v", compress.Round(longitude, 6.0), compress.Round((float64(i)*0.00001+2.123456)*-1, 6.0))
			t.FailNow()
		}

		if compress.Round(latitude, 6.0) != compress.Round(float64(i)*0.00001+1.98765, 6.0) {
			t.Logf("latitude error: %v != %v", compress.Round(latitude, 6.0), compress.Round(float64(i)*0.00001+1.98765, 6.0))
			t.FailNow()
		}
	}
}
