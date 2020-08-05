package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"math/big"
	)

// This wrapper file is used for easily mocking in unit test.
// It is also a good convention to create a wrapper when calling database
type hasSNDerivatorFunc func(stateDB *statedb.StateDB, tokenID common.Hash, snd []byte) (bool, error)
type hasSerialNumberFunc func(stateDB *statedb.StateDB, coinID common.Hash, serialNumber []byte, shardID byte) (bool, error)
type hasCommitmentFunc func(stateDB *statedb.StateDB, tokenID common.Hash, commitment []byte, shardID byte) (bool, error)
type hasCommitmentIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) (bool, error)
type getCommitmentByIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, commitmentIndex uint64, shardID byte) ([]byte, error)
type getCommitmentIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, commitment []byte, shardID byte) (*big.Int, error)
type getCommitmentLengthFunc func(stateDB *statedb.StateDB, tokenID common.Hash, shardID byte) (*big.Int, error)
type hasOnetimeAddressFunc func(stateDB *statedb.StateDB, tokenID common.Hash, ota []byte) (bool, error)
type hasOTACoinIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, index uint64, shardID byte) (bool, error)
type getOTACoinByIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, index uint64, shardID byte) ([]byte, error)
type getOTACoinIndexFunc func(stateDB *statedb.StateDB, tokenID common.Hash, ota []byte) (*big.Int, error)
type getOTACoinLengthFunc func(stateDB *statedb.StateDB, tokenID common.Hash, shardID byte) (*big.Int, error)
type privacyTokenIDExistedFunc func(stateDB *statedb.StateDB, tokenID common.Hash) bool
type getAllBridgeTokensFunc func(stateDB *statedb.StateDB) ([]byte, error)
type isBridgeTokenExistedByTypeFunc func(stateDB *statedb.StateDB, incTokenID common.Hash, isCentralized bool) (bool, error)
type txDbWrapper struct {
	hasSNDerivator  hasSNDerivatorFunc
	hasSerialNumber hasSerialNumberFunc

	// Mainly used in ver 1
	hasCommitment hasCommitmentFunc
	hasCommitmentIndex hasCommitmentIndexFunc
	getCommitmentByIndex getCommitmentByIndexFunc
	getCommitmentIndex getCommitmentIndexFunc
	getCommitmentLength getCommitmentLengthFunc

	// Mainly used in ver 2
	//hasOnetimeAddress hasOnetimeAddressFunc
	//hasOTACoinIndex hasOTACoinIndexFunc
	//getOTACoinByIndex getOTACoinByIndexFunc
	//getOTACoinIndex getOTACoinIndexFunc
	//getOTACoinLength getOTACoinLengthFunc

	// Used in token
	privacyTokenIDExisted privacyTokenIDExistedFunc
	getAllBridgeTokens getAllBridgeTokensFunc
	isBridgeTokenExistedByType isBridgeTokenExistedByTypeFunc
}
var txDatabaseWrapper = NewTxDbWrapper()
func NewTxDbWrapper() txDbWrapper {
	return txDbWrapper{
		hasSNDerivator:  statedb.HasSNDerivator,
		hasSerialNumber: statedb.HasSerialNumber,
		hasCommitment: statedb.HasCommitment,
		hasCommitmentIndex: statedb.HasCommitmentIndex,
		getCommitmentByIndex: statedb.GetCommitmentByIndex,
		getCommitmentIndex: statedb.GetCommitmentIndex,
		getCommitmentLength: statedb.GetCommitmentLength,
		//hasOnetimeAddress: statedb.HasOnetimeAddress,
		//hasOTACoinIndex: statedb.HasOTACoinIndex,
		//getOTACoinByIndex: statedb.GetOTACoinByIndex,
		//getOTACoinIndex: statedb.GetOTACoinIndex,
		//getOTACoinLength: statedb.GetOTACoinLength,
		privacyTokenIDExisted: statedb.PrivacyTokenIDExisted,
		getAllBridgeTokens: statedb.GetAllBridgeTokens,
		isBridgeTokenExistedByType: statedb.IsBridgeTokenExistedByType,
	}
}
