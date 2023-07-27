package util

import (
	"errors"
	"math"
	"os"
)

func DegreesToRadians(a float64) float64 { return math.Pi * a / 180.0 }

func RadiansToDegrees(a float64) float64 { return 180.0 * a / math.Pi }

func Pythagoras(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2.0) + math.Pow(y2-y1, 2.0))
}

func Round(value float64) float64 {
	var roundOn = 0.5
	var places = 7.0

	var round float64
	pow := math.Pow(10, places)
	digit := pow * value
	_, div := math.Modf(digit)

	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return round / pow
}

func Equal(pointA, pointB [2]float64) (equal bool) {
	if Round(pointA[0]) != Round(pointB[0]) || Round(pointA[1]) != Round(pointB[1]) {
		return false
	}

	return true
}

// ChangeRootDir
//
// # English:
//
// It looks for the directory defined in dirToSearch and if it doesn't find it, it goes up the main directory one level
//
//	Input:
//	  dirToSearch: name of the directory where to save the downloaded and processed files
//
// # Português:
//
// Procura pelo diretório definido em dirToSearch e caso não encontre, sobe o diretório principal em um nível
//
//	Entrada:
//	  dirToSearch: nome do diretório onde salvar os arquivos baixados e processados
func ChangeRootDir(dirToSearch string) (err error) {
	// change main dir to open 'commonFiles' folder
	var dir []os.DirEntry
	var safetyCounter = 5
	for {
		pass := false
		dir, err = os.ReadDir("./")
		if err != nil {
			return
		}

		for k := range dir {
			if dir[k].Name() == dirToSearch {
				pass = true
				break
			}
		}

		if pass {
			break
		}

		err = os.Chdir("../")
		if err != nil {
			return
		}

		safetyCounter -= 1
		if safetyCounter == 0 {
			err = errors.New("for security reasons, the limit is five fetch interactions")
			return
		}
	}

	return
}
