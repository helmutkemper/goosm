package opentriedb

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

// searchTestFunc uses binary search to find and return the smallest index i
// in [0, n) at which f(i) is true, assuming that on the range [0, n),
// f(i) == true implies f(i+1) == true. That is, Search requires that
// f is false for some (possibly empty) prefix of the input range [0, n)
// and then true for the (possibly empty) remainder; Search returns
// the first true index. If there is no such index, Search returns n.
// (Note that the "not found" return value is not -1 as in, for instance,
// strings.Index.)
// Search calls f(i) only for i in the range [0, n).
//
// A common use of Search is to find the index i for a value x in
// a sorted, indexable data structure such as an array or slice.
// In this case, the argument f, typically a closure, captures the value
// to be searched for, and how the data structure is indexed and
// ordered.
//
// For instance, given a slice data sorted in ascending order,
// the call Search(len(data), func(i int) bool { return data[i] >= 23 })
// returns the smallest index i such that data[i] >= 23. If the caller
// wants to find whether 23 is in the slice, it must test data[i] == 23
// separately.
//
// Searching data sorted in descending order would use the <=
// operator instead of the >= operator.
//
// To complete the example above, the following code tries to find the value
// x in an integer slice data sorted in ascending order:
//
//	x := 23
//	i := sort.Search(len(data), func(i int) bool { return data[i] >= x })
//	if i < len(data) && data[i] == x {
//		// x is present at data[i]
//	} else {
//		// x is not present in data,
//		// but i is the index where it would be inserted.
//	}
//
// As a more whimsical example, this program guesses your number:
//
//	func GuessingGame() {
//		var s string
//		fmt.Printf("Pick an integer from 0 to 100.\n")
//		answer := sort.Search(100, func(i int) bool {
//			fmt.Printf("Is your number <= %d? ", i)
//			fmt.Scanf("%s", &s)
//			return s != "" && s[0] == 'y'
//		})
//		fmt.Printf("Your number is %d.\n", answer)
//	}
var searchTestFunc func(n int, f func(int64) bool) int

// searchUI32TestFunc searches for x in a sorted slice of uint32 and returns the index
// as specified by Search. The return value is the index to insert x if x is
// not present (it could be len(a)).
// The slice must be sorted in ascending order.
var searchUI32TestFunc func(a []uint32, x uint32) int

// randDataTestFunc this function assembles an array with random values, of length size,
// with values between 0 and max included
var randDataTestFunc func(size int, max int) []uint32

// randSourceTest source of random numbers initialized with time.Now()
var randSourceTest *rand.Rand

func TestMain(m *testing.M) {

	// randSourceTest source of random numbers initialized with time.Now()
	randSourceTest = rand.New(rand.NewSource(time.Now().UnixNano()))

	// randDataTestFunc this function assembles an array with random values, of length size,
	// with values between 0 and max included
	randDataTestFunc = func(size int, max int) (data []uint32) {
		data = make([]uint32, 0)
		max += 1
		for {
			n := uint32(randSourceTest.Intn(max))
			if n == 0 {
				continue
			}
			pass := true
			for _, v := range data {
				if v == n {
					pass = false
					break
				}
			}
			if pass == true {
				data = append(data, n)
			}

			if len(data) == size {
				break
			}
		}

		return
	}

	// Search uses binary search to find and return the smallest index i
	// in [0, n) at which f(i) is true, assuming that on the range [0, n),
	// f(i) == true implies f(i+1) == true. That is, Search requires that
	// f is false for some (possibly empty) prefix of the input range [0, n)
	// and then true for the (possibly empty) remainder; Search returns
	// the first true index. If there is no such index, Search returns n.
	// (Note that the "not found" return value is not -1 as in, for instance,
	// strings.Index.)
	// Search calls f(i) only for i in the range [0, n).
	//
	// A common use of Search is to find the index i for a value x in
	// a sorted, indexable data structure such as an array or slice.
	// In this case, the argument f, typically a closure, captures the value
	// to be searched for, and how the data structure is indexed and
	// ordered.
	//
	// For instance, given a slice data sorted in ascending order,
	// the call Search(len(data), func(i int) bool { return data[i] >= 23 })
	// returns the smallest index i such that data[i] >= 23. If the caller
	// wants to find whether 23 is in the slice, it must test data[i] == 23
	// separately.
	//
	// Searching data sorted in descending order would use the <=
	// operator instead of the >= operator.
	//
	// To complete the example above, the following code tries to find the value
	// x in an integer slice data sorted in ascending order:
	//
	//	x := 23
	//	i := sort.Search(len(data), func(i int) bool { return data[i] >= x })
	//	if i < len(data) && data[i] == x {
	//		// x is present at data[i]
	//	} else {
	//		// x is not present in data,
	//		// but i is the index where it would be inserted.
	//	}
	//
	// As a more whimsical example, this program guesses your number:
	//
	//	func GuessingGame() {
	//		var s string
	//		fmt.Printf("Pick an integer from 0 to 100.\n")
	//		answer := sort.Search(100, func(i int) bool {
	//			fmt.Printf("Is your number <= %d? ", i)
	//			fmt.Scanf("%s", &s)
	//			return s != "" && s[0] == 'y'
	//		})
	//		fmt.Printf("Your number is %d.\n", answer)
	//	}
	searchTestFunc = func(n int, f func(int64) bool) int {
		// Define f(-1) == false and f(n) == true.
		// Invariant: f(i-1) == false, f(j) == true.
		i, j := int64(0), int64(n)
		for i < j {
			h := int64(uint64(i+j) >> 1) // avoid overflow when computing h
			// i ≤ h < j
			if !f(h) {
				i = h + 1 // preserves f(i-1) == false
			} else {
				j = h // preserves f(j) == true
			}
		}
		// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
		return int(i)
	}

	// SearchUI32 searches for x in a sorted slice of uint32 and returns the index
	// as specified by Search. The return value is the index to insert x if x is
	// not present (it could be len(a)).
	// The slice must be sorted in ascending order.
	searchUI32TestFunc = func(a []uint32, x uint32) int {
		return searchTestFunc(len(a), func(i int64) bool { return a[i] >= x })
	}

	result := m.Run()
	os.Exit(result)
}

func TestMainLine_InsertPiece(t *testing.T) {
	const (
		kDataLineLength = 64
		linePieceSize   = 4
	)

	var file *os.File
	var err error

	_ = os.Remove("./test.bin")
	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		t.Logf("os.OpenFile().err: %v", err)
		t.FailNow()
	}

	t.Cleanup(
		func() {
			_ = file.Close()
			_ = os.Remove("./test.bin")
		},
	)

	ml := new(mainLine)
	err = ml.init(file, &TextUTF16{})
	if err != nil {
		t.Fatalf("ml.init().error: %v", err)
	}

	// Condições:
	//
	// | condição da linha                                                 | próxima piece                              |
	// |-------------------------------------------------------------------|--------------------------------------------|
	// | [ ] a linha não tem nada, length = 0                              | terminal                                   |
	// | [ ] a linha tem dados, porém, há espaço, length < kDataLineLength | terminal                                   |
	// | [ ] a linha tem dados e está lotada, length == kDataLineLength    | terminal                                   |
	// | [ ] a linha não tem nada, length = 0                              | não terminal                               |
	// | [ ] a linha tem dados, porém, há espaço, length < kDataLineLength | não terminal                               |
	// | [ ] a linha tem dados e está lotada, length == kDataLineLength    | não terminal                               |

	// Respostas:
	//
	// | address       | terminal       | keySlot       | nextAddressToInsert       | addressToUpdate       | i         |
	// |---------------|----------------|---------------|---------------------------|-----------------------|-----------|
	// |             0 |           true |             1 |                        -1 |                     0 |         0 |
	// |             0 |           true |             2 |                        -1 |                     0 |         1 |
	// |             0 |           true |            63 |                        -1 |                     0 |        62 |
	// |             0 |           true |            64 |                        -1 |                     0 |        63 |
	// |             0 |           true |             1 |                        -1 |                   844 |        64 |
	// |             0 |           true |             2 |                        -1 |                   844 |        65 |
	// |             0 |           true |             3 |                        -1 |                   844 |        66 |
	// |             0 |           true |             4 |                        -1 |                   844 |        67 |
	// |             0 |          false |             1 |                       844 |                     0 |         0 |
	// |             0 |          false |             2 |                      1688 |                     0 |         1 |
	// |             0 |          false |            63 |                     53172 |                     0 |        62 |
	// |             0 |          false |            64 |                     54016 |                     0 |        63 |
	// |             0 |          false |             1 |                     55704 |                 54860 |        64 |
	// |             0 |          false |             2 |                     54860 |                 55704 |        65 |
	// |             0 |          false |             3 |                     57392 |                 54860 |        66 |
	// |             0 |          false |             4 |                     58236 |                 54860 |        67 |
	// |             0 |          false |             5 |                     59080 |                 54860 |        68 |

	var nextAddressToInsert, addressToUpdate int64
	var keySlot int
	for i := 0; i != kDataLineLength+5; i += 1 {
		piece := make([]byte, linePieceSize)
		binary.BigEndian.PutUint32(piece, uint32(i))
		address := ml.getAddressStart()
		keySlot, nextAddressToInsert, addressToUpdate, err = ml.insertPiece(address, piece, true)
		if err != nil {
			t.Fatalf("ml.insertPiece().error: %v", err)
		}

		if i >= kDataLineLength {
			var found bool
			var addressLine int64
			var flagData flag
			found, addressLine, flagData, err = ml.searchPieceInDataLine(address, piece)
			if err != nil {
				t.Fatalf("ml.searchPieceInDataLine().error: %v", err)
			}

			_ = found
			_ = addressLine
			_ = flagData
		}
	}

	_ = addressToUpdate
	_ = nextAddressToInsert
	_ = keySlot
}

func TestMainLine_searchSuccessiveApproximations(t *testing.T) {
	const (
		kDataLineLength = 64
	)

	var file *os.File
	var err error

	_ = os.Remove("./test.bin")
	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		t.Logf("os.OpenFile().err: %v", err)
		t.FailNow()
	}

	t.Cleanup(
		func() {
			_ = file.Close()
			_ = os.Remove("./test.bin")
		},
	)

	ml := new(mainLine)
	err = ml.init(file, &TextUTF16{})
	if err != nil {
		t.Fatalf("ml.init().error: %v", err)
	}

	dataInserted := make([]uint32, 0)

	dataToInsert := randDataTestFunc(kDataLineLength*5, kDataLineLength*5)
	var nextAddressListLength int
	var nextAddressList []int64
	var length int
	var found bool
	var index int
	var piece []byte
	for _, value := range dataToInsert {
		nextAddressListLength, nextAddressList, length, err = ml.readLengthInDataCluster(0x00)
		if err != nil {
			t.Fatalf("ml.readLengthInDataCluster().error: %v", err)
		}

		// if index == 0, the cluster is full and will need to add one more row of data
		if length != 0 && length%kDataLineLength == 0 {
			var nextAddressToInsert int64
			nextAddressToInsert, err = ml.writeNewDataLineInCluster(nextAddressList[nextAddressListLength-1])
			if err != nil {
				t.Fatalf("ml.writeNewDataLineInCluster().error: %v", err)
			}

			nextAddressList = append(nextAddressList, nextAddressToInsert)
			nextAddressListLength += 1
		}

		found, index, _, err = ml.searchSuccessiveApproximations(nextAddressList[0], int64(value), length)
		if err != nil {
			t.Fatalf("ml.searchSuccessiveApproximations().error: %v", err)
		}

		if found {
			t.Fatal("the piece has not yet been inserted and should not have been found")
		}

		err = ml.shiftRightSlotInCluster(index, nextAddressListLength, nextAddressList, &length)
		if err != nil {
			t.Fatalf("ml.shiftRightSlotInCluster().error: %v", err)
		}

		piece = make([]byte, kRuneByteSize)
		binary.LittleEndian.PutUint32(piece, value)
		err = ml.writeSlotInCluster(nextAddressList[0], piece, -1, flagComplete, index)
		if err != nil {
			t.Fatalf("ml.writeSlotInCluster().error: %v", err)
		}

		if index >= length {
			length, err = ml.readLength(nextAddressList[nextAddressListLength-1])
			if err != nil {
				t.Fatalf("ml.readLength().error: %v", err)
			}
			err = ml.writeLength(nextAddressList[nextAddressListLength-1], length+1)
			if err != nil {
				t.Fatalf("ml.writeLength().error: %v", err)
			}
		}

		if cap(dataInserted) < index {
			log.Print("")
		}
		dataInserted = insert(dataInserted, index, value)
	}

	// checks each entered data if order is correct
	for k := 0; k != len(dataInserted); k += 1 {
		piece, err = ml.readPieceInDataCluster(nextAddressList[0], k)
		if err != nil {
			t.Fatalf("ml.readPiece().error: %v", err)
		}
		ui32 := binary.LittleEndian.Uint32(piece)
		if ui32 != dataInserted[k] {
			t.Fatalf("data error! [%v]: %v != %v", k, ui32, dataInserted[k])
		}
	}
}

// TestMainLine_shiftRightSlotInDataLine creates a list of numbers between 0 and 63, included, in random order and then
// tests successive approximations to test the shiftRightSlotInDataLine() function
func TestMainLine_shiftRightSlotInDataLine(t *testing.T) {
	const (
		kDataLineLength = 64
	)

	//todo: fazer um teste onde index = ultimo elemento na DataLine

	var file *os.File
	var err error

	_ = os.Remove("./test.bin")
	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		t.Logf("os.OpenFile().err: %v", err)
		t.FailNow()
	}

	t.Cleanup(
		func() {
			_ = file.Close()
			_ = os.Remove("./test.bin")
		},
	)

	ml := new(mainLine)
	err = ml.init(file, &TextUTF16{})
	if err != nil {
		t.Fatalf("ml.init().error: %v", err)
	}

	dataInserted := make([]uint32, 0)

	dataToInsert := randDataTestFunc(kDataLineLength+1, kDataLineLength+1)
	for k, x := range dataToInsert {
		// search for successive approximations
		i := searchUI32TestFunc(dataInserted, x)

		if k == kDataLineLength {
			i = 0
			x = 0
		}

		// If the number is greater than the highest number, numbers greater than the new number are shifted to the right.
		var flagLength = false
		if i < len(dataInserted) {
			var overflow bool
			var overflowPiece []byte
			var overflowAddressLine int64
			var overflowFlagData flag
			overflow, overflowPiece, overflowAddressLine, overflowFlagData, err = ml.shiftRightSlotInDataLine(0x00, i)
			if err != nil {
				t.Fatalf("ml.shiftRightSlotInDataLine().error: %v", err)
			}

			if k != kDataLineLength && overflow == true {
				t.Fatal("flag overflow fail")
			}

			if k != kDataLineLength && overflowPiece != nil {
				t.Fatal("piece overflow fail")
			}

			if k != kDataLineLength && overflowAddressLine != 0x00 {
				t.Fatal("addressLine overflow fail")
			}

			if k != kDataLineLength && overflowFlagData != flagNotSet {
				t.Fatal("flag overflow fail")
			}

			flagLength = true
		}

		dataInserted = insert(dataInserted, i, x)

		piece := make([]byte, 4)
		binary.LittleEndian.PutUint32(piece, x)
		err = ml.writeSlot(0x00, piece, -1, flagComplete, i)
		if err != nil {
			t.Fatalf("ml.insertPiece().error: %v", err)
		}

		var length int
		length, err = ml.readLength(0x00)
		if err != nil {
			t.Fatalf("ml.writeLength().error: %v", err)
		}
		if flagLength {
			length -= 1 // the data will be written over a duplicated data by shiftRightSlotInDataLine()
		}
		err = ml.writeLength(0x00, length+1)
		if err != nil {
			t.Fatalf("ml.writeLength().error: %v", err)
		}

		// checks each entered data if order is correct
		var l int
		if l = len(dataInserted); l > kDataLineLength {
			l = kDataLineLength
		}
		for k := 0; k != l; k += 1 {
			piece, err = ml.readPiece(0x00, k)
			if err != nil {
				t.Fatalf("ml.readPiece().error: %v", err)
			}
			ui32 := binary.LittleEndian.Uint32(piece)
			if ui32 != dataInserted[k] {
				t.Fatalf("data error! [%v][%v]: %v != %v", i, k, ui32, dataInserted[k])
			}
		}
	}
}

//func TestMainLine_shiftRightSlotInCluster_1(t *testing.T) {
//	var file *os.File
//	var err error
//
//	_ = os.Remove("./test.bin")
//	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
//	if err != nil {
//		t.Logf("os.OpenFile().err: %v", err)
//		t.FailNow()
//	}
//
//	t.Cleanup(
//		func() {
//			_ = file.Close()
//			_ = os.Remove("./test.bin")
//		},
//	)
//
//	ml := new(mainLine)
//	err = ml.init(file)
//	if err != nil {
//		t.Fatalf("ml.init().error: %v", err)
//	}
//
//	for i := 0; i != kDataLineLength*4+1; i += 1 {
//		piece := []byte{byte(i), byte(i), byte(i), byte(i)}
//		_, _, _, err = ml.insertPiece(0x00, piece, true)
//		if err != nil {
//			t.Fatalf("ml.insertPiece().error: %v", err)
//		}
//	}
//
//	var overflow bool
//	var overflowPiece []byte
//	var overflowAddressLine int64
//	var overflowFlagData flag
//	err = ml.shiftRightSlotInCluster(0x00, kDataLineLength*2+7)
//	if err != nil {
//		t.Fatalf("ml.shiftRightSlotInDataLine().error: %v", err)
//	}
//
//	if overflow != true {
//		t.Fatal("flag overflow fail")
//	}
//
//	if !bytes.Equal(overflowPiece, []byte{byte(0xff), byte(0xff), byte(0xff), byte(0xff)}) {
//		t.Fatal("piece overflow fail")
//	}
//
//	if overflowAddressLine != -1 {
//		t.Fatal("addressLine overflow fail")
//	}
//
//	if overflowFlagData != 0x02 {
//		t.Fatal("flag overflow fail")
//	}
//
//	piece, err := ml.readPiece(0x00, 0)
//	if err != nil {
//		t.Fatalf("ml.insertPiece().error: %v", err)
//	}
//
//	if !bytes.Equal(piece, []byte{0x00, 0x00, 0x00, 0x00}) {
//		t.Fatal("piece copy fail")
//	}
//
//	for i := 0; i != kDataLineLength*4; i += 1 {
//		piece, err = ml.readPieceInDataCluster(0x00, i)
//		if err != nil {
//			t.Fatalf("ml.insertPiece().error: %v", err)
//		}
//
//		value := i
//		if i > kDataLineLength*2+7 {
//			value -= 1
//		}
//
//		if !bytes.Equal(piece, []byte{byte(value), byte(value), byte(value), byte(value)}) {
//			t.Fatalf("piece copy fail: %v", i)
//		}
//	}
//}

func insert(a []uint32, index int, value uint32) []uint32 {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func TestMainLine_getPieceByIndexInDataCluster(t *testing.T) {
	const (
		kDataLineLength = 64
	)

	var file *os.File
	var err error

	_ = os.Remove("./test.bin")
	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		t.Logf("os.OpenFile().err: %v", err)
		t.FailNow()
	}

	t.Cleanup(
		func() {
			_ = file.Close()
			_ = os.Remove("./test.bin")
		},
	)

	ml := new(mainLine)
	err = ml.init(file, &TextUTF16{})
	if err != nil {
		t.Fatalf("ml.init().error: %v", err)
	}

	for i := 0; i != kDataLineLength*4; i += 1 {
		piece := []byte{byte(i), byte(i), byte(i), byte(i)}
		_, _, _, err = ml.insertPiece(0x00, piece, true)
		if err != nil {
			t.Fatalf("ml.insertPiece().error: %v", err)
		}
	}

	var found bool
	var piece []byte
	var addressLine int64
	var flagData flag
	for i := 0; i != kDataLineLength*4; i += 1 {
		found, piece, addressLine, flagData, err = ml.getPieceByIndexInDataCluster(0x00, i)
		if err != nil {
			t.Fatalf("ml.getPieceByIndexInDataCluster().error: %v", err)
		}

		if !found {
			t.Fatal("flag found fail")
		}

		if !bytes.Equal(piece, []byte{byte(i), byte(i), byte(i), byte(i)}) {
			t.Fatal("piece fail")
		}

		if addressLine != -1 {
			t.Fatal("addressLine fail")
		}

		if flagData != 0x02 {
			t.Fatal("flag fail")
		}
	}
}

func TestMainLine_InsertKey(t *testing.T) {
	const (
		linePieceSize = 4
	)

	var file *os.File
	var err error

	//_ = os.Remove("./test.bin")
	file, err = os.OpenFile("./test.bin", os.O_CREATE|os.O_RDWR, fs.ModePerm)
	if err != nil {
		t.Logf("os.OpenFile().err: %v", err)
		t.FailNow()
	}

	t.Cleanup(
		func() {
			_ = file.Close()
			//_ = os.Remove("./test.bin")
		},
	)

	ml := new(mainLine)
	err = ml.init(file, &TextUTF16{})
	if err != nil {
		t.Fatalf("ml.init().error: %v", err)
	}

	dataKey := ml.stringToKey("Hello world!")
	var nextAddressToInsert, addressToUpdate int64
	var keySlot int
	address := ml.getAddressStart()
	for {
		piece := make([]byte, linePieceSize)
		copy(piece, dataKey)
		dataKey = dataKey[linePieceSize:]

		keyLength := len(dataKey)

		keySlot, nextAddressToInsert, addressToUpdate, err = ml.insertPiece(address, piece, keyLength == 0)
		if err != nil {
			t.Fatalf("ml.insertPiece().error: %v", err)
		}

		address = nextAddressToInsert

		if keyLength == 0 {
			break
		}
	}

	_ = addressToUpdate
	_ = nextAddressToInsert
	_ = keySlot
}
