package privacy

import (
	"fmt"
	"testing"
)

func TestAES(t*testing.T){
	aes := new(AES)
	aes.Key = []byte{203, 156, 98, 88, 68, 146, 29, 168, 247, 221, 65, 60, 47, 173, 232, 41, 149, 186, 230, 10, 56, 116, 35, 155, 230, 105, 172, 130, 174, 123, 63, 199}

	msg, err := aes.decrypt([]byte{181, 246, 203, 70, 77, 51, 181, 85, 75, 214, 231, 230, 91, 205, 149, 119, 214, 223})
	if err != nil{
		fmt.Printf("ERR: %v\n", err)
	}

	fmt.Printf("Message: %v\n", msg)

	//res := PedCom.G[0].Derive(big.NewInt(2), big.NewInt(10))
	//fmt.Printf("res: %v\n", res.Compress())

}



