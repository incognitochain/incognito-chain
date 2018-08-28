package wallet

import (
	"io/ioutil"
	"strings"
	"path/filepath"
)

func NewWordList(language string) []string {
	absPath, err := filepath.Abs("./" + language + ".txt")
	if err != nil {
		Logger.log.Error(err)
	}
	englishBytes, err := ioutil.ReadFile(absPath)
	english := string(englishBytes)
	if err != nil {
		Logger.log.Error(err)
	}
	var englishWords = strings.Split(strings.TrimSpace(english), "\n")
	return englishWords
}
