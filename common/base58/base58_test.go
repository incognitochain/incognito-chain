package base58

import (
	"encoding/base64"
	"fmt"
	rand2 "math/rand"
	"testing"
	"time"
)

func RandBytes(length int) []byte {
	seed := time.Now().UnixNano()
	b := make([]byte, length)
	reader := rand2.New(rand2.NewSource(int64(seed)))

	for n := 0; n < length; {
		read, err := reader.Read(b[n:])
		if err != nil {
			fmt.Printf("[PRIVACY LOG] Rand byte error : %v\n", err)
			return nil
		}
		n += read
	}
	return b
}

var data = RandBytes(100000)

func TestFastEndcode(t *testing.T) {
	fmt.Println(base64.StdEncoding.EncodeToString(data))
	r := Encode(data)
	fmt.Println(r)
}

func TestFastDecode(t *testing.T) {
	fmt.Println(base64.StdEncoding.EncodeToString(data))
	r := Encode(data)
	fmt.Println(r)
	d, _ := Decode(r)
	fmt.Println(base64.StdEncoding.EncodeToString(d))
}
