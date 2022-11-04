package goosm

import (
	"io"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func BenchmarkBinary(b *testing.B) {
	//7.961.992.696 - /Users/kemper/GolandProjects/pbf/node.tmp.bin
	var err error
	var tmpFilePath = "../node.tmp.bin"
	var pbfProcess = PbfProcess{}
	var tmpFile *os.File
	tmpFile, err = os.OpenFile(tmpFilePath, os.O_RDONLY, fs.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := tmpFile.Close()
		if err != nil {
			log.Printf("error closing file %v - %v", tmpFilePath, err)
		}
	}()

	data := make([]byte, 8)
	index := int64(0)
	for {
		_, err = tmpFile.ReadAt(data, index*8*3)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if err == io.EOF {
			break
		}

		index++
	}

	idList := []int64{273316, 257392564, 257576887, 257585146, 257745745, 257747636, 258640948, 259480239, 263918118, 266967762, 266983481, 267191395, 267193001, 267194355, 267451995, 269222058, 269273961, 269712058, 273210744, 277285545, 10073738084, 10073064071, 10073038998, 10069785384, 10068685377, 10067759486}
	rand.Seed(time.Now().Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeID := rand.Int63n(int64(len(idList) - 1))
		_, _, err = pbfProcess.nodeSearchInTmpFile(tmpFile, index, nodeID)
		if err != nil {
			log.Fatal(err)
		}
	}
}
