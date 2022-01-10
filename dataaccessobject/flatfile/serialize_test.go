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

	id, _ := ff.Append([]byte(genStr("1")))
	fmt.Println(id)
	ff.Append([]byte(genStr("2")))
	ff.Append([]byte(genStr("3")))
	ff.Append([]byte(genStr("4")))
	ff.Append([]byte(genStr("5")))
	ff.Append([]byte(genStr("6")))

	str := []byte(genStr("7"))
	id, _ = ff.Append(str)
	fmt.Println(id)
	fmt.Println("str", str)

	str2, err := ff.Read(id)
	fmt.Println("read", str2, err)

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
