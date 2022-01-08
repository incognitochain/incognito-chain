package flatfile

import (
	"fmt"
	"testing"
	"time"
)

func TestNewFlatFile(t *testing.T) {

	ff, _ := NewFlatFile("./", 3)
	var genStr = func(s string) string {
		res := ""
		for i := 0; i < 10; i++ {
			res += s
		}
		return res
	}

	ff.Append([]byte(genStr("1")))
	ff.Append([]byte(genStr("2")))
	ff.Append([]byte(genStr("3")))
	ff.Append([]byte(genStr("4")))
	ff.Append([]byte(genStr("5")))
	ff.Append([]byte(genStr("6")))
	ff.Append([]byte(genStr("7")))

	c, e, _ := ff.ReadRecently()
	for {
		select {
		case msg := <-c:
			fmt.Println(msg)
		case msg := <-e:
			fmt.Println(msg)
		}
		time.Sleep(time.Second)
	}

}
