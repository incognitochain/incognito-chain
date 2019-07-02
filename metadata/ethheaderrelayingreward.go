package metadata

import (
	"bytes"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
)

type ETHHeaderRelayingReward struct {
	RequestedTxID common.Hash
	MetadataBase
}

func NewETHHeaderRelayingReward(
	requestedTxID common.Hash,
	metaType int,
) *ETHHeaderRelayingReward {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	ethHeaderRelayingReward := &ETHHeaderRelayingReward{
		RequestedTxID: requestedTxID,
	}
	ethHeaderRelayingReward.MetadataBase = metadataBase
	return ethHeaderRelayingReward
}

func (e *ETHHeaderRelayingReward) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (e *ETHHeaderRelayingReward) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, true, nil
}

func (e *ETHHeaderRelayingReward) ValidateMetadataByItself() bool {
	if e.Type != ETHHeaderRelayingRewardMeta {
		return false
	}
	return true
}

func (e *ETHHeaderRelayingReward) Hash() *common.Hash {
	record := e.RequestedTxID.String()
	record += e.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (e *ETHHeaderRelayingReward) CalculateSize() uint64 {
	return calculateSize(e)
}

func (e *ETHHeaderRelayingReward) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	uniqETHTxsUsed [][]byte,
) (bool, error) {
	idx := -1
	for i, txInBlock := range txsInBlock {
		if txsUsed[i] > 0 ||
			txInBlock.GetMetadataType() != ETHHeaderRelayingMeta ||
			!bytes.Equal(e.RequestedTxID[:], txInBlock.Hash()[:]) {
			continue
		}
		ethHeaderRelayingRaw := txInBlock.GetMetadata()
		ethHeaderRelaying, ok := ethHeaderRelayingRaw.(*ETHHeaderRelaying)
		if !ok {
			continue
		}
		lc := bcr.GetLightEthereum().GetLightChain()
		_, err := lc.ValidateHeaderChain(ethHeaderRelaying.ETHHeaders, 0)
		if err != nil {
			fmt.Printf("ETH header relaying validation failed: %v", err)
			continue
		}

		_, pk, amount, _ := tx.GetTransferData()
		reward := txInBlock.GetTxFee() + uint64(len(ethHeaderRelaying.ETHHeaders)*1)
		if !bytes.Equal(ethHeaderRelaying.RelayerAddress.Pk[:], pk[:]) ||
			reward != amount {
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingRequest tx found for IssuingResponse tx %s", tx.Hash().String())
	}
	txsUsed[idx] = 1
	return true, nil
}
