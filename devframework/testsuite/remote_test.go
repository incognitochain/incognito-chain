package testsuite

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/devframework"
)

func Test_Remote(t *testing.T) {
	rm := devframework.NewRPCClient("23.234.324.2:8000")
	fmt.Println(rm)
}
