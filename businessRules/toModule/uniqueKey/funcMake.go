package uniqueKey

import (
	"math/rand"
	"time"
)

func (e *UniqueKey) Make() (uniqueKey string) {
	var uniqueKeyBase = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
		"w", "x", "y", "z", "A", "B", "C", "D", "E", "F",
		"G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
		"Q", "R", "S", "T", "W", "X", "Y", "Z", "0", "1",
		"2", "3", "4", "5", "6", "7", "8", "9", "!", "@",
		"#", "$", "%", "&", "*", "(", ")", "_", "=", "-",
		"+", "[", "]", "{", "}", "|", "/", "?", "<", ">",
		";", ":",
	}
	var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i != 50; i += 1 {
		uniqueKey += uniqueKeyBase[randomGenerator.Intn(82-1)]
	}

	return
}
