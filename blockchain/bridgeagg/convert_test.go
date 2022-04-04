package bridgeagg

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/stretchr/testify/suite"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("Init logger")
	return
}()

type ConvertTestSuite struct {
	suite.Suite
	producerState  State
	processorState State

	sdb *statedb.StateDB
	env *stateEnvironment
}

func (c *ConvertTestSuite) SetupTest() {

}
