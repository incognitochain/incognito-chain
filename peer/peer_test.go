package peer

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()
