package opentriedb

import "encoding/binary"

type TextUTF16 struct {
	data *[]byte
}

// SetData defines the pointer to the data to be archived
//
//	Warning: GetNextPiece will destroy the data because it uses a pointer to save memory
//
//go:inline
func (e *TextUTF16) SetData(data *[]byte) {
	e.data = data
}

// GetNextPiece returns a piece to be filed in the DataLine
//
//	Warning: GetNextPiece will destroy the data because it uses a pointer to save memory
//	Note: to learn more about DataLine or piece, see the doc.go file
//
//go:inline
func (e *TextUTF16) GetNextPiece() (piece []byte) {
	piece = make([]byte, 4)
	copy(piece, *e.data)
	*e.data = (*e.data)[4:]
	return
}

// GetPieceLength returns the amount of bytes needed to archive a piece
//
//	Note: to learn more about DataLine or piece, see the doc.go file
//
//go:inline
func (e *TextUTF16) GetPieceLength() (length int64) {
	return 4
}

// GetDataLineLength returns the size of the DataLine optimized for the chosen type
//
//	Note: to learn more about DataLine or piece, see the doc.go file
//
//go:inline
func (e *TextUTF16) GetDataLineLength() (length int) {
	return 64
}

// ToPiece turns a number into piece
//
//	Note: to learn more about DataLine or piece, see the doc.go file
//
//go:inline
func (e *TextUTF16) ToPiece(number int64) (piece []byte) {
	piece = make([]byte, 4)
	binary.LittleEndian.PutUint32(piece, uint32(number))
	return
}

// ToNumber turns a piece into a number
//
//	Note: to learn more about DataLine or piece, see the doc.go file
//
//go:inline
func (e *TextUTF16) ToNumber(piece []byte) (number int64) {
	// Warning: uint32 must first be converted to int32 and then converted to int64
	// To better understand, see converting numbers from base 2 to base 10 and how negative numbers are represented
	return int64(int32(binary.LittleEndian.Uint32(piece)))
}
