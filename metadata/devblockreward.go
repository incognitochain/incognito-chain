package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type DevRewardInfo struct {
	DevReward map[common.Hash]uint64
}

func BuildInstForDevReward(reward map[common.Hash]uint64, devAddress string) ([]string, error) {

	devRewardInfo := DevRewardInfo{
		DevReward: reward,
	}

	contentStr, err := json.Marshal(devRewardInfo)
	if err != nil {
		return nil, err
	}

	keyWalletDevAccount, err := wallet.Base58CheckDeserialize(devAddress)
	if err != nil {
		Logger.log.Error(NewMetadataTxError(WrongDevPaymentAddressError, err))
		return nil, err
	}
	returnedInst := []string{
		strconv.Itoa(DevRewardRequestMeta),
		strconv.Itoa(int(common.GetShardIDFromLastByte(keyWalletDevAccount.KeySet.PaymentAddress.Pk[common.PublicKeySize-1]))),
		"devRewardInst",
		string(contentStr),
	}

	return returnedInst, nil
}

func NewDevRewardInfoFromStr(inst string) (*DevRewardInfo, error) {
	Ins := &DevRewardInfo{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}
