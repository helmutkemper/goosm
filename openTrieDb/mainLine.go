package opentriedb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"runtime"
)

// PieceMaker allows changing the behavior of the way of archiving data and optimizing memory consumption for each data
// type, numerical, ascii text, utf8, utf16, etc.
type PieceMaker interface {
	// SetData defines the pointer to the data to be archived
	//  Warning: GetNextPiece will destroy the data because it uses a pointer to save memory
	SetData(data *[]byte)

	// GetNextPiece returns a piece to be filed in the DataLine
	//  Warning: GetNextPiece will destroy the data because it uses a pointer to save memory
	//  Note: to learn more about DataLine or piece, see the doc.go file
	GetNextPiece() (piece []byte)

	// GetPieceLength returns the amount of bytes needed to archive a piece
	//  Note: to learn more about DataLine or piece, see the doc.go file
	GetPieceLength() (length int64)

	// GetDataLineLength returns the size of the DataLine optimized for the chosen type
	//  Note: to learn more about DataLine or piece, see the doc.go file
	GetDataLineLength() (length int)

	// ToPiece turns a number into piece
	//  Note: to learn more about DataLine or piece, see the doc.go file
	ToPiece(number int64) (piece []byte)

	// ToNumber turns a piece into a number
	//  Note: to learn more about DataLine or piece, see the doc.go file
	ToNumber(piece []byte) (number int64)
}

type flag byte

const (
	flagNotSet flag = iota
	flagContinue
	flagComplete
	flagAmbiguous
	flagItSelf
)

type mainLine struct {
	address int64
	// address of the next line to be written to the file
	nextAddressLine int64
	file            *os.File

	pieceMaker PieceMaker
}

//go:inline
func (e *mainLine) incNextAddressLine() {
	e.nextAddressLine += e.formulaNextKey(e.pieceMaker.GetDataLineLength())
}

//go:inline
func (e *mainLine) getNextAddressLine() (nextAddressLine int64) {
	e.incNextAddressLine()
	return e.nextAddressLine
}

//go:inline
func (e *mainLine) getAddressStart() (address int64) {
	return headerLength
}

//go:inline
func (e *mainLine) formulaNextKey(key int) (address int64) {
	//              length  + next address + (piece + address line + flag)*key
	address = uint32Size + uint64Size + (e.pieceMaker.GetPieceLength()+uint64Size+byteSize)*int64(key)
	return
}

//go:inline
func (e *mainLine) init(file *os.File, pieceMaker PieceMaker) (err error) {
	e.file = file
	e.pieceMaker = pieceMaker

	err = e.writeBlankLine(e.getAddressStart())
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	e.address = headerLength
	e.nextAddressLine = headerLength
	return
}

//func (e *mainLine) insert(key, value []byte) (err error) {
//	lenPiece := len(key)
//	address := e.getAddressStart()
//
//	for {
//		piece := make([]byte, linePieceSize)
//		copy(piece, key)
//		key = key[linePieceSize:]
//		lenPiece -= linePieceSize
//
//		terminal := lenPiece == 0
//
//		var found bool
//		var addressLine int64
//		var flagData flag
//		found, addressLine, flagData, err = e.searchPieceInDataLine(address, piece)
//		if found && !terminal {
//			address = addressLine
//			continue
//		}
//
//		// !found
//		// found && terminal
//		//   trocar flagComplete para flagContComp
//
//		e.insertPiece(address, piece, terminal)
//	}
//}

//go:inline
func (e *mainLine) insertPiece(address int64, piece []byte, terminal bool) (keySlot int, nextAddressToInsert int64, addressToUpdate int64, err error) {

	var length int
	length, err = e.readLength(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	if length == e.pieceMaker.GetDataLineLength() {
		nextAddressToInsert, err = e.readNextAddress(address)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		if nextAddressToInsert == 0 {
			_, err = e.writeNewDataLineInCluster(address)
			err = e.writeBlankLine(nextAddressToInsert)
			if err != nil {
				err = errors.Join(e.errorColector(), err)
				return
			}
		}

		return e.insertPiece(nextAddressToInsert, piece, terminal)
	}

	flagData := flagContinue
	nextAddressToInsert = -1
	if terminal {
		flagData = flagComplete
	} else {
		// reserves a new line for the next piece
		nextAddressToInsert = e.getNextAddressLine() // todo: isso deveria sair daqui
		err = e.writeBlankLine(nextAddressToInsert)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}
	}

	err = e.writeLength(address, length+1)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeSlot(address, piece, nextAddressToInsert, flagData, length)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	return length, nextAddressToInsert, address, err
}

//go:inline
func (e *mainLine) readNextAddress(address int64) (nextAddress int64, err error) {
	data := make([]byte, uint64Size)
	_, err = e.file.ReadAt(data, address+uint32Size) //address+ length size
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	nextAddressUi64 := binary.LittleEndian.Uint64(data)
	nextAddress = int64(nextAddressUi64)
	return
}

//go:inline
func (e *mainLine) writeNextAddress(address int64, nextAddress int64) (err error) {
	data := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(data, uint64(nextAddress))
	_, err = e.file.WriteAt(data, address+uint32Size) //address+ length size
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	return
}

//go:inline
func (e *mainLine) writeLength(address int64, length int) (err error) {
	data := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(data, uint32(length))
	_, err = e.file.WriteAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
	}
	return
}

//go:inline
func (e *mainLine) readPieceInDataCluster(address int64, index int) (piece []byte, err error) {
	var nextAddressList []int64
	_, nextAddressList, _, err = e.readLengthInDataCluster(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	key := index / e.pieceMaker.GetDataLineLength()
	index = index % e.pieceMaker.GetDataLineLength()

	piece = make([]byte, e.pieceMaker.GetPieceLength())
	_, err = e.file.ReadAt(piece, nextAddressList[key]+e.formulaNextKey(index))
	return
}

//go:inline
func (e *mainLine) readLengthInDataCluster(address int64) (nextAddressListLength int, nextAddressList []int64, length int, err error) {
	//todo: mudar nextAddressListLength
	nextAddressListLength = 1
	nextAddressList = []int64{address}

	var lengthAtual int
	for {
		lengthAtual, err = e.readLength(address)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}
		length += lengthAtual

		var nextAddress int64
		nextAddress, err = e.readNextAddress(address)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		if nextAddress == 0 {
			break
		}

		nextAddressListLength += 1
		nextAddressList = append(nextAddressList, nextAddress)
		address = nextAddress
	}

	return
}

//go:inline
func (e *mainLine) readLength(address int64) (length int, err error) {
	data := make([]byte, uint32Size)
	_, err = e.file.ReadAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	lengthUi32 := binary.LittleEndian.Uint32(data)
	length = int(lengthUi32)
	return
}

//go:inline
func (e *mainLine) readPiece(address int64, index int) (piece []byte, err error) {
	piece = make([]byte, e.pieceMaker.GetPieceLength())
	_, err = e.file.ReadAt(piece, address+e.formulaNextKey(index))
	return
}

//go:inline
func (e *mainLine) writePiece(address int64, piece []byte, index int) (err error) {
	_, err = e.file.WriteAt(piece, address+e.formulaNextKey(index))
	return
}

//go:inline
func (e *mainLine) readAddressLine(address int64, key int) (addressLine int64, err error) {
	data := make([]byte, uint64Size)
	_, err = e.file.ReadAt(data, address+e.pieceMaker.GetPieceLength()+e.formulaNextKey(key))
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	addressLine = int64(binary.LittleEndian.Uint64(data))
	return
}

//go:inline
func (e *mainLine) writeAddressLine(address int64, addressLine int64, key int) (err error) {
	data := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(data, uint64(addressLine))
	_, err = e.file.WriteAt(data, address+e.pieceMaker.GetPieceLength()+e.formulaNextKey(key))
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	return
}

//go:inline
func (e *mainLine) writeFlag(address int64, flag flag, key int) (err error) {
	_, err = e.file.WriteAt([]byte{(byte)(flag)}, address+e.pieceMaker.GetPieceLength()+uint64Size+e.formulaNextKey(key))
	return
}

//go:inline
func (e *mainLine) readFlag(address int64, key int) (flagData flag, err error) {
	data := make([]byte, byteSize)
	_, err = e.file.ReadAt(data, address+e.pieceMaker.GetPieceLength()+uint64Size+e.formulaNextKey(key))
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	flagData = (flag)(data[0])
	return
}

// writeSlot: Write piece, addressLine, and flag at slot index x, inside address addr.
// Due to the function's unique responsibility, it does not change how much data there is in lineData.
//
//	Input:
//	  address: lineData address
//	  piece: data to be inserted
//	  addressLine: address of the next addressLine or of the final data
//	  flag: archived data type indicator byte
//	  index: slot index into the addressLine
//
//go:inline
func (e *mainLine) writeSlot(address int64, piece []byte, addressLine int64, flag flag, index int) (err error) {
	err = e.writePiece(address, piece, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeAddressLine(address, addressLine, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeFlag(address, flag, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
	}

	return
}

//go:inline
func (e *mainLine) readSlot(address int64, index int) (piece []byte, addressLine int64, flagData flag, err error) {
	if index >= e.pieceMaker.GetDataLineLength() {
		err = fmt.Errorf("index (%v) must be less than %v", index, e.pieceMaker.GetDataLineLength())
		err = errors.Join(e.errorColector(), err)
		return
	}

	piece, err = e.readPiece(address, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	addressLine, err = e.readAddressLine(address, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	flagData, err = e.readFlag(address, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
	}

	return
}

//go:inline
func (e *mainLine) writeBlankLine(address int64) (err error) {
	err = e.writeLength(address, 0)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeNextAddress(address, 0)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeBlankKeys(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
	}

	return
}

//go:inline
func (e *mainLine) writeBlankKeys(address int64) (err error) {
	piece := make([]byte, e.pieceMaker.GetPieceLength())
	for key := 0; key != e.pieceMaker.GetDataLineLength(); key += 1 {
		err = e.writeSlot(address, piece, 0, flagNotSet, key)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}
	}
	return
}

//go:inline
func (e *mainLine) shiftRightSlotInCluster(index int, nextAddressListLength int, nextAddressList []int64, length *int) (err error) {
	if index >= *length || *length == 0 {
		return
	}

	key := index / e.pieceMaker.GetDataLineLength()
	index = index % e.pieceMaker.GetDataLineLength()

	for i := nextAddressListLength - 1; i != key-1; i -= 1 {
		indexChange := 0
		if i == key {
			indexChange = index
		}

		var overflow bool
		var piece []byte
		var addressLine int64
		var flagData flag

		overflow, piece, addressLine, flagData, err = e.shiftRightSlotInDataLine(nextAddressList[i], indexChange)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		if overflow {
			err = e.writeSlot(nextAddressList[i+1], piece, addressLine, flagData, 0)
			if err != nil {
				err = errors.Join(e.errorColector(), err)
				return
			}
		}

		if i == key {
			return
		}
	}

	return
}

//go:inline
func (e *mainLine) writeNewDataLineInCluster(address int64) (nextAddressToInsert int64, err error) {
	nextAddressToInsert = e.getNextAddressLine()
	err = e.writeBlankLine(nextAddressToInsert)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeNextAddress(address, nextAddressToInsert)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	return
}

//go:inline
func (e *mainLine) shiftRightSlotInDataLine(address int64, index int) (overflow bool, overflowPiece []byte, overflowAddressLine int64, overflowFlagData flag, err error) {
	var length int
	length, err = e.readLength(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	if length != e.pieceMaker.GetDataLineLength() {
		err = e.writeLength(address, length+1)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		if length == 0 {
			return
		}
	}

	if length == e.pieceMaker.GetDataLineLength() {
		overflow = true
		overflowPiece, overflowAddressLine, overflowFlagData, err = e.readSlot(address, e.pieceMaker.GetDataLineLength()-1)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		// index of the last element of the line, it is pushed out with the overflow flag
		if index == e.pieceMaker.GetDataLineLength()-1 {
			overflow = true
			return
		}

		// como length está no limite, o ultimo dado é jogado para fora
		length -= 1
	}

	var piece []byte
	var addressLine int64
	var flagData flag
	for i := length - 1; ; i -= 1 {
		piece, addressLine, flagData, err = e.readSlot(address, i)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		err = e.writeSlot(address, piece, addressLine, flagData, i+1) // se i == 63, cria nova data Line e i = 0
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		if i == index {
			return
		}
	}
}

//go:inline
func (e *mainLine) getPieceByIndexInDataCluster(address int64, index int) (found bool, piece []byte, addressLine int64, flagData flag, err error) {
	var length int
	var nextAddressList []int64
	_, nextAddressList, length, err = e.readLengthInDataCluster(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	if index > length-1 {
		return
	}

	key := index / e.pieceMaker.GetDataLineLength()
	index = index % e.pieceMaker.GetDataLineLength()
	address = nextAddressList[key]

	found = true
	piece, addressLine, flagData, err = e.readSlot(address, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	return
}

//go:inline
func (e *mainLine) searchPieceInDataLine(address int64, piece []byte) (found bool, addressLine int64, flagData flag, err error) {
	var length int
	length, err = e.readLength(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return false, 0, flagNotSet, err
	}

	var pieceRead []byte
	for key := 0; key != length; key += 1 {
		pieceRead, addressLine, flagData, err = e.readSlot(address, key)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return false, 0, flagNotSet, err
		}

		if bytes.Equal(piece, pieceRead) {
			return true, addressLine, flagData, nil
		}
	}

	if length == e.pieceMaker.GetDataLineLength() {
		var nextAddress int64
		nextAddress, err = e.readNextAddress(address)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return false, 0, flagNotSet, err
		}

		return e.searchPieceInDataLine(nextAddress, piece)
	}

	return false, 0, flagNotSet, nil
}

//func (e *mainLine) updateDataLine(address int64, piece []byte, terminal bool) (err error) {
//	var length int
//	length, err = e.readLength(address)
//	if err != nil {
//		err = errors.Join(e.errorColector(), err)
//		return
//	}
//
//	var nextAddress int64
//	nextAddress, err = e.readNextAddress(address)
//	if err != nil {
//		err = errors.Join(e.errorColector(), err)
//		return
//	}
//
//	var found bool
//	var addressLineRead int64
//	var flagDataRead flag
//	found, addressLineRead, flagDataRead, err = e.searchPieceInDataLine(address, piece)
//	if err != nil {
//		err = errors.Join(e.errorColector(), err)
//		return
//	}
//
//	// não existe
//	//   adicionar
//	//     a linha está lotada
//	//       newLine, next addr = newLine addr, gravar em slot 0
//	//     a linha não está lotada
//	//       length+=1, next addr 0x00, gravar em slot len+1
//	// existente piece é terminal
//	//   nova piece é terminal
//	//   nova piece não é terminal
//	// existente piece não é terminal
//	//   nova piece é terminal
//	//   nova piece não é terminal
//
//}

// writeHeader writes the header of the binary file, with the format version and author data of the original code
//
//	Warning: the header file does not store the total amount of data, as there would be a value always being written to
//	         the same place on the SSD and this will reduce the useful life of the SSD
//
//go:inline
func (e *mainLine) writeHeader(address int64) (nextAddress int64, err error) {
	// write version
	data := []byte("openTrieDB V0.0.1\n\n")
	length := len(data)
	_, err = e.file.WriteAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	address += int64(length)

	// write contact info
	data = []byte("openTrieDB by Helmut Kemper - helmut.kemper@gmail.com\n")
	length = len(data)
	_, err = e.file.WriteAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	address += int64(length)

	data = []byte("+55 81999268744\n")
	length = len(data)
	_, err = e.file.WriteAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	address += int64(length)

	data = []byte("https://github.com/helmutkemper\n\n")
	length = len(data)
	_, err = e.file.WriteAt(data, address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}
	address += int64(length)

	if address != headerLength {
		err = errors.New("the header has been changed and the size is incorrect")
		err = errors.Join(e.errorColector(), err)
		return
	}

	return address, nil
}

//go:inline
func (e *mainLine) errorColector() (err error) {
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		f := runtime.FuncForPC(pc)
		funcName := f.Name()
		err = fmt.Errorf("\nfunc: %v [line: %v]\n%v", funcName, line, file)
		return
	}

	err = errors.New("errorColector error")
	return
}

//go:inline
func (e *mainLine) stringToKey(keyString string) (key []byte) {
	key = make([]byte, len(keyString)*4)
	for k, letter := range keyString {
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, uint32(letter)) //todo: aqui
		copy(key[k*4:], data)
	}

	return
}

//go:inline
func (e *mainLine) writeSlotInCluster(address int64, piece []byte, addressLine int64, flag flag, index int) (err error) {
	var nextAddressList []int64
	_, nextAddressList, _, err = e.readLengthInDataCluster(address)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	key := index / e.pieceMaker.GetDataLineLength()
	index = index % e.pieceMaker.GetDataLineLength()

	err = e.writePiece(nextAddressList[key], piece, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeAddressLine(nextAddressList[key], addressLine, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	err = e.writeFlag(nextAddressList[key], flag, index)
	if err != nil {
		err = errors.Join(e.errorColector(), err)
		return
	}

	return
}

//go:inline
func (e *mainLine) searchSuccessiveApproximations(address int64, value int64, length int) (found bool, index int, piece []byte, err error) {

	var function = func(index int) (found, greaterThanOrEqualTo bool, piece []byte, err error) {
		piece, err = e.readPieceInDataCluster(address, index)
		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}

		dataUI := e.pieceMaker.ToNumber(piece)
		return dataUI == value, dataUI >= value, piece, err
	}

	greaterThanOrEqualTo := false
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := 0, length
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if found, greaterThanOrEqualTo, piece, err = function(h); !greaterThanOrEqualTo {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}

		if err != nil {
			err = errors.Join(e.errorColector(), err)
			return
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return found, i, piece, err
}
