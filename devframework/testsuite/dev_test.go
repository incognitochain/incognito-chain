package testsuite

import (
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/devframework"
)

func Test_Devnode(t *testing.T) {
	blockchain.ReadKey()
	blockchain.SetupParam()
	chainParam := &blockchain.ChainTest2Param
	devframework.NewIncognitoNode("devnode", chainParam, "45.56.115.6:9330", true)

	select {}
}
