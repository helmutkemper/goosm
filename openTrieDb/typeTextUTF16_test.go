package opentriedb

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestTextUTF16_SetData(t *testing.T) {
	stringToKey := func(keyString string) (key []byte) {
		key = make([]byte, len(keyString)*4)
		for k, letter := range keyString {
			data := make([]byte, 4)
			binary.LittleEndian.PutUint32(data, uint32(letter))
			copy(key[k*4:], data)
		}

		return
	}

	text := "the quick brown fox jumps over the lazy dog"
	response := [][]byte{
		{byte('t'), 0, 0, 0}, {byte('h'), 0, 0, 0}, {byte('e'), 0, 0, 0}, {byte(' '), 0, 0, 0}, {byte('q'), 0, 0, 0},
		{byte('u'), 0, 0, 0}, {byte('i'), 0, 0, 0}, {byte('c'), 0, 0, 0}, {byte('k'), 0, 0, 0}, {byte(' '), 0, 0, 0},
		{byte('b'), 0, 0, 0}, {byte('r'), 0, 0, 0}, {byte('o'), 0, 0, 0}, {byte('w'), 0, 0, 0}, {byte('n'), 0, 0, 0},
		{byte(' '), 0, 0, 0}, {byte('f'), 0, 0, 0}, {byte('o'), 0, 0, 0}, {byte('x'), 0, 0, 0}, {byte(' '), 0, 0, 0},
		{byte('j'), 0, 0, 0}, {byte('u'), 0, 0, 0}, {byte('m'), 0, 0, 0}, {byte('p'), 0, 0, 0}, {byte('s'), 0, 0, 0},
		{byte(' '), 0, 0, 0}, {byte('o'), 0, 0, 0}, {byte('v'), 0, 0, 0}, {byte('e'), 0, 0, 0}, {byte('r'), 0, 0, 0},
		{byte(' '), 0, 0, 0}, {byte('t'), 0, 0, 0}, {byte('h'), 0, 0, 0}, {byte('e'), 0, 0, 0}, {byte(' '), 0, 0, 0},
		{byte('l'), 0, 0, 0}, {byte('a'), 0, 0, 0}, {byte('z'), 0, 0, 0}, {byte('y'), 0, 0, 0}, {byte(' '), 0, 0, 0},
		{byte('d'), 0, 0, 0}, {byte('o'), 0, 0, 0}, {byte('g'), 0, 0, 0},
	}
	data := stringToKey(text)
	p := new(TextUTF16)
	p.SetData(&data)

	for _, resp := range response {
		piece := p.GetNextPiece()
		if !bytes.Equal(piece, resp) {
			t.Fatalf("TextUTF16.GetNextPiece().error: %X != %X", piece, resp)
		}
	}

	if p.GetPieceLength() != 4 {
		t.Fatal("TextUTF16.GetPieceLength().error: the value must be 4 for UTF16")
	}

	piece := make([]byte, 4)
	for i := int64(math.MinInt32); i != math.MaxInt32; i += 1 {
		binary.LittleEndian.PutUint32(piece, uint32(i))
		if !bytes.Equal(piece, p.ToPiece(i)) {
			t.Fatalf("TextUTF16.ToPiece().error: %X != %X", piece, p.ToPiece(i))
		}

		if p.ToNumber(piece) != i {
			t.Fatalf("TextUTF16.ToNumber().error: %X != %X", i, p.ToNumber(piece))
		}
	}
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
