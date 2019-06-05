package metadata

import (
	"encoding/json"

	// "errors"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type DevRewardInfo struct {
	DevReward uint64
}

func BuildInstForDevReward(reward uint64) ([]string, error) {

	devRewardInfo := DevRewardInfo{
		DevReward: reward,
	}

	contentStr, err := json.Marshal(devRewardInfo)
	if err != nil {
		return nil, err
	}

	devAddress, err := privacy.NewPaymentAddressFromString(common.DevAddress)

	returnedInst := []string{
		strconv.Itoa(DevRewardRequestMeta),
		strconv.Itoa(int(common.GetShardIDFromLastByte(devAddress.Bytes()[len(devAddress.Bytes())-1]))),
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
