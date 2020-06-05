package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []coin.Coin
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
	OutputCoin       []coin.Coin
}
type ContentCrossShardTokenPrivacyData struct {
	OutputCoin     []coin.Coin
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
func (contentCrossShardTokenPrivacyData *ContentCrossShardTokenPrivacyData) UnmarshalJSON(data []byte) error {
	type Alias ContentCrossShardTokenPrivacyData
	temp := &struct {
		OutputCoin []string
		*Alias
	}{
		Alias: (*Alias)(contentCrossShardTokenPrivacyData),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		Logger.log.Error("UnmarshalJSON ContentCrossShardTokenPrivacyData", err)
		return err
	}
	outputCoinList, err := coin.ParseCoinsStr(temp.OutputCoin)
	if err != nil {
		Logger.log.Error("UnmarshalJSON Cannot parse crossOutputCoins", err)
		return err
	}
	contentCrossShardTokenPrivacyData.OutputCoin = outputCoinList
	return nil
}

func (crossOutputCoin CrossOutputCoin) Hash() common.Hash {
	res := []byte{}
	res = append(res, crossOutputCoin.BlockHash.GetBytes()...)
	for _, coins := range crossOutputCoin.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	return common.HashH(res)
}
func (crossOutputCoin *CrossOutputCoin) UnmarshalJSON(data []byte) error {
	type Alias CrossOutputCoin
	temp := &struct {
		OutputCoin []string
		*Alias
	}{
		Alias: (*Alias)(crossOutputCoin),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		Logger.log.Error("UnmarshalJSON CrossOutputCoin", err)
		return err
	}
	outputCoinList, err := coin.ParseCoinsStr(temp.OutputCoin)
	if err != nil {
		Logger.log.Error("UnmarshalJSON Cannot parse CrossOutputCoin", err)
		return err
	}
	crossOutputCoin.OutputCoin = outputCoinList
	return nil
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
func (crossTransaction *CrossTransaction) UnmarshalJSON(data []byte) error {
	type Alias CrossTransaction
	temp := &struct {
		OutputCoin []string
		*Alias
	}{
		Alias: (*Alias)(crossTransaction),
	}
	if err := json.Unmarshal(data, temp); err != nil {
		Logger.log.Error("UnmarshalJSON CrossTransaction", string(data))
		return err
	}
	outputCoinList, err := coin.ParseCoinsStr(temp.OutputCoin)
	if err != nil {
		Logger.log.Error("UnmarshalJSON Cannot parse CrossTransaction", err)
		return err
	}
	crossTransaction.OutputCoin = outputCoinList
	return nil
}

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
func (crossShardBlock *CrossShardBlock) VerifyCrossShardBlock(blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committees); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(crossShardBlock, crossShardBlock.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
