package metadata

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type ReturnStakingMetadata struct {
	MetadataBase
	TxID          string
	StakerAddress privacy.PaymentAddress
}

func NewReturnStaking(
	txID string,
	producerAddress privacy.PaymentAddress,
	metaType int,
) *ReturnStakingMetadata {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &ReturnStakingMetadata{
		TxID:          txID,
		StakerAddress: producerAddress,
		MetadataBase:  metadataBase,
	}
}

func (sbsRes ReturnStakingMetadata) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes ReturnStakingMetadata) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, stateDB *statedb.StateDB) (bool, error) {
	stakingTx := bcr.GetStakingTx(shardID)
	for key, value := range stakingTx {
		committeePublicKey := incognitokey.CommitteePublicKey{}
		err := committeePublicKey.FromString(key)
		if err != nil {
			return false, err
		}
		if reflect.DeepEqual(sbsRes.StakerAddress.Pk, committeePublicKey.IncPubKey) && (sbsRes.TxID == value) {
			autoStakingList := bcr.GetAutoStakingList()
			if autoStakingList[key] {
				return false, errors.New("Can not return staking amount for candidate, who want to restaking.")
			}
			return true, nil
		}
	}
	return false, errors.New("Can not find any staking information of this publickey")
}

func (sbsRes ReturnStakingMetadata) ValidateSanityData(bcr BlockchainRetriever, txr Transaction, beaconHeight uint64) (bool, bool, error) {
	if len(sbsRes.StakerAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if len(sbsRes.StakerAddress.Tk) == 0 {
		return false, false, errors.New("Wrong request info's producer address")
	}
	if sbsRes.TxID == "" {
		return false, false, errors.New("Wrong request info's Tx staking")
	}
	return false, true, nil
}

func (sbsRes ReturnStakingMetadata) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes ReturnStakingMetadata) Hash() *common.Hash {
	record := sbsRes.StakerAddress.String()
	record += sbsRes.TxID

	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}
