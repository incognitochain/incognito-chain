package metadata

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type MintDelegationRewardMetadata struct {
	MetadataBase
	TxID            string
	ReceiverAddress privacy.PaymentAddress
	RewardAmount    uint64
	SharedRandom    []byte `json:"SharedRandom,omitempty"`
}

func NewMintDelegationReward(
	txID string,
	receiverAddr privacy.PaymentAddress,
	rewardAmount uint64,
	metaType int,
) *MintDelegationRewardMetadata {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &MintDelegationRewardMetadata{
		TxID:            txID,
		ReceiverAddress: receiverAddr,
		RewardAmount:    rewardAmount,
		MetadataBase:    metadataBase,
	}
}

func (sbsRes MintDelegationRewardMetadata) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, stateDB *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes MintDelegationRewardMetadata) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

// pk: 32, tk: 32
func (sbsRes MintDelegationRewardMetadata) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {

	if len(sbsRes.ReceiverAddress.Pk) != common.PublicKeySize {
		return false, false, errors.New("Wrong request info's receiver address")
	}

	if len(sbsRes.ReceiverAddress.Tk) != common.TransmissionKeySize {
		return false, false, errors.New("Wrong request info's receiver address")
	}

	if sbsRes.TxID == "" {
		return false, false, errors.New("Wrong request info's Tx request reward")
	}

	_, err := common.Hash{}.NewHashFromStr(sbsRes.TxID)
	if err != nil {
		return false, false, errors.New("Wrong request info's Tx request reward hash")
	}

	if sbsRes.RewardAmount == 0 {
		return false, false, errors.New("Cant mint zero reward")
	}

	return false, true, nil
}

func (sbsRes MintDelegationRewardMetadata) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes MintDelegationRewardMetadata) Hash() *common.Hash {
	record := sbsRes.ReceiverAddress.String()
	record += "-" + sbsRes.TxID
	record += "-" + strconv.FormatUint(sbsRes.RewardAmount, 10)
	if sbsRes.SharedRandom != nil && len(sbsRes.SharedRandom) > 0 {
		record += "-" + string(sbsRes.SharedRandom)
	}
	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}

func (sbsRes *MintDelegationRewardMetadata) SetSharedRandom(r []byte) {
	sbsRes.SharedRandom = r
}
