package flatfile

import (
	"fmt"
	"testing"
	"time"
)

func TestFlatFileManager_ReadFromIndex(t *testing.T) {
	ff, err := NewFlatFile("/tmp/123", 10)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		ff.Append([]byte(fmt.Sprint(i)))
	}

	c, e, _ := ff.ReadFromIndex(0)
	go func() {
		<-e
		panic(err)
	}()
	for {
		data := <-c
		fmt.Println(string(data))
		if len(data) == 0 {
			time.Sleep(time.Second)
			return
		}
	}
}
