package main

import (
	"goosm/goosm"
	"goosm/plugin/mongodb"
	"testing"
	"time"
)

// BenchmarkInsertID_1-8   	    1224	   1107270 ns/op
func BenchmarkInsertID_1(b *testing.B) {
	var err error
	var timeout = 10 * time.Second

	dbNode := &mongodb.DbNode{}
	_, err = dbNode.New("mongodb://127.0.0.1:27016/", "osm", "node", timeout)
	if err != nil {
		b.Logf("%v", err)
		b.FailNow()
	}

	b.ResetTimer()

	var node goosm.Node
	for i := 0; i < b.N; i++ {
		node, err = dbNode.GetById(1)
		if err != nil {
			b.Logf("%v", err)
			b.FailNow()
		}
	}
	_ = node
}

// BenchmarkFindNodeById-8   	          146314	      9390 ns/op
func BenchmarkInsertID_1736017358(b *testing.B) {
	var err error
	var timeout = 10 * time.Second

	dbNode := &mongodb.DbNode{}
	_, err = dbNode.New("mongodb://127.0.0.1:27016/", "osm", "node", timeout)
	if err != nil {
		b.Logf("%v", err)
		b.FailNow()
	}

	b.ResetTimer()

	var node goosm.Node
	for i := 0; i < b.N; i++ {
		node, err = dbNode.GetById(1736017358)
		if err != nil {
			b.Logf("%v", err)
			b.FailNow()
		}
	}
	_ = node
}
