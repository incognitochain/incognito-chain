package instructionsprocessor

import (
	"encoding/binary"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []privacy.OutputCoin
}

type CrossTokenPrivacyData struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossShardTokenPrivacyData
}
type CrossTransaction struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossShardTokenPrivacyData
	OutputCoin       []privacy.OutputCoin
}
type ContentCrossShardTokenPrivacyData struct {
	OutputCoin     []privacy.OutputCoin
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}
type CrossShardTokenPrivacyMetaData struct {
	TokenID        common.Hash
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}

func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Bytes() []byte {
	res := []byte{}
	for _, item := range contentCrossShardTokenPrivacyData.OutputCoin {
		res = append(res, item.Bytes()...)
	}
	res = append(res, contentCrossShardTokenPrivacyData.PropertyID.GetBytes()...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertyName)...)
	res = append(res, []byte(contentCrossShardTokenPrivacyData.PropertySymbol)...)
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, uint32(contentCrossShardTokenPrivacyData.Type))
	res = append(res, typeBytes...)
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(amountBytes, uint32(contentCrossShardTokenPrivacyData.Amount))
	res = append(res, amountBytes...)
	if contentCrossShardTokenPrivacyData.Mintable {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	return res
}
func (contentCrossShardTokenPrivacyData ContentCrossShardTokenPrivacyData) Hash() common.Hash {
	return common.HashH(contentCrossShardTokenPrivacyData.Bytes())
}
func (crossOutputCoin CrossOutputCoin) Hash() common.Hash {
	res := []byte{}
	res = append(res, crossOutputCoin.BlockHash.GetBytes()...)
	for _, coins := range crossOutputCoin.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	return common.HashH(res)
}
func (crossTransaction CrossTransaction) Bytes() []byte {
	res := []byte{}
	res = append(res, crossTransaction.BlockHash.GetBytes()...)
	for _, coins := range crossTransaction.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	for _, coins := range crossTransaction.TokenPrivacyData {
		res = append(res, coins.Bytes()...)
	}
	return res
}
func (crossTransaction CrossTransaction) Hash() common.Hash {
	return common.HashH(crossTransaction.Bytes())
}

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
// func (crossShardBlock *CrossShardBlock) VerifyCrossShardBlock(blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
// 	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committees); err != nil {
// 		return  err
// 	}
// 	if ok := VerifyCrossShardBlockUTXO(crossShardBlock, crossShardBlock.MerklePathShard); !ok {
// 		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
// 	}
// 	return nil
// }
