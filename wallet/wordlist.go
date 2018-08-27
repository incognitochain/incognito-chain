package wallet

import (
	"io/ioutil"
	"strings"
)

func NewWordList(language string) []string {
	englishBytes, err := ioutil.ReadFile(language + ".txt")
	english := string(englishBytes)
	if err != nil {
		Logger.log.Error(err)
	}
	//checksum := crc32.ChecksumIEEE([]byte(english))
	//c1dbd296 := fmt.Sprintf("%x", checksum)
	//if c1dbd296 != "c1dbd296" {
	//	panic("english.txt checksum invalid")
	//}
	var englishWords = strings.Split(strings.TrimSpace(english), "\n")
	return englishWords
}
