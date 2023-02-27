package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type ReturnBeaconStakingMetadata struct {
	MetadataBase
	TxIDs         []string
	StakerAddress privacy.PaymentAddress
	SharedRandom  []byte `json:"SharedRandom,omitempty"`
}

func NewReturnBeaconStaking(txIDs []string, producerAddress privacy.PaymentAddress, metaType int) *ReturnBeaconStakingMetadata {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ReturnBeaconStakingMetadata{
		TxIDs:         txIDs,
		StakerAddress: producerAddress,
		MetadataBase:  metadataBase,
	}
}

func (sbsRes ReturnBeaconStakingMetadata) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes ReturnBeaconStakingMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

// pk: 32, tk: 32
func (sbsRes ReturnBeaconStakingMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {

	if len(sbsRes.StakerAddress.Pk) != common.PublicKeySize {
		return false, false, errors.New("Wrong request info's producer address")
	}

	if len(sbsRes.StakerAddress.Tk) != common.TransmissionKeySize {
		return false, false, errors.New("Wrong request info's producer address")
	}

	if len(sbsRes.TxIDs) == 0 {
		return false, false, errors.New("Wrong request info's Tx staking")
	}
	for _, txID := range sbsRes.TxIDs {
		_, err := common.Hash{}.NewHashFromStr(txID)
		if err != nil {
			return false, false, errors.New("Wrong request info's Tx staking hash")
		}
	}
	return false, true, nil
}

func (sbsRes ReturnBeaconStakingMetadata) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes ReturnBeaconStakingMetadata) Hash() *common.Hash {
	record := sbsRes.StakerAddress.String()
	for id, v := range sbsRes.TxIDs {
		if id > 0 {
			record += "-"
		}
		record += v
	}
	if sbsRes.SharedRandom != nil && len(sbsRes.SharedRandom) > 0 {
		record += string(sbsRes.SharedRandom)
	}
	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (sbsRes *ReturnBeaconStakingMetadata) SetSharedRandom(r []byte) {
	sbsRes.SharedRandom = r
}

func (sbsRes *ReturnBeaconStakingMetadata) ToCompactBytes() ([]byte, error) {
	return toCompactBytes(sbsRes)
}
