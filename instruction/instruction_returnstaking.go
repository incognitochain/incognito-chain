package instruction

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//ReturnStakeIns :
// format: "return", "key1,key2,key3", "2", "1231231,312312321,12312321", "100,100,100,100"
type ReturnStakeIns struct {
	PublicKeys       []string
	PublicKeysStruct []incognitokey.CommitteePublicKey
	ShardID          byte
	StakingTXIDs     []string
	StakingTxHashes  []common.Hash
	PercentReturns   []uint
}

func NewReturnStakeInsWithValue(
	publicKeys []string,
	sID byte,
	txStake []string,
	pReturn []uint,
) *ReturnStakeIns {
	return &ReturnStakeIns{
		PublicKeys:     publicKeys,
		ShardID:        sID,
		StakingTXIDs:   txStake,
		PercentReturns: pReturn,
	}
}

func NewReturnStakeIns() *ReturnStakeIns {
	return &ReturnStakeIns{}
}

func (rsI *ReturnStakeIns) SetShardID(sID byte) error {
	rsI.ShardID = sID
	return nil
}

func (rsI *ReturnStakeIns) SetPublicKeys(publicKeys []string) (*ReturnStakeIns, error) {
	rsI.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	rsI.PublicKeysStruct = publicKeyStructs
	return rsI, nil
}

func (rsI *ReturnStakeIns) SetStakingTXIDs(txIDs []string) (*ReturnStakeIns, error) {
	rsI.StakingTXIDs = txIDs
	return rsI, nil
}

func (rsI *ReturnStakeIns) SetPercentReturns(percentReturns []uint) error {
	rsI.PercentReturns = percentReturns
	return nil
}

func (rsI *ReturnStakeIns) GetType() string {
	return RETURN_ACTION
}

func (rsI *ReturnStakeIns) GetShardID() byte {
	return rsI.ShardID
}

func (rsI *ReturnStakeIns) GetPercentReturns() []uint {
	return rsI.PercentReturns
}

func (rsI *ReturnStakeIns) GetStakingTX() []string {
	return rsI.StakingTXIDs
}

func (rsI *ReturnStakeIns) GetPublicKey() []string {
	return rsI.PublicKeys
}

func (rsI *ReturnStakeIns) ToString() []string {
	returnStakeInsStr := []string{RETURN_ACTION}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.PublicKeys, SPLITTER))
	returnStakeInsStr = append(returnStakeInsStr, strconv.Itoa(int(rsI.ShardID)))
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.StakingTXIDs, SPLITTER))
	percentReturnsStr := make([]string, len(rsI.PercentReturns))
	for i, v := range rsI.PercentReturns {
		percentReturnsStr[i] = strconv.Itoa(int(v))
	}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(percentReturnsStr, SPLITTER))
	return returnStakeInsStr
}

func ValidateAndImportReturnStakingInstructionFromString(instruction []string) (*ReturnStakeIns, error) {
	if err := ValidateReturnStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnStakingInstructionFromString(instruction)
}

func ImportReturnStakingInstructionFromString(instruction []string) (*ReturnStakeIns, error) {
	returnStakingIns := NewReturnStakeIns()
	var err error
	returnStakingIns, err = returnStakingIns.SetPublicKeys(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}

	shardID, err := strconv.Atoi(instruction[2])
	if err != nil {
		return nil, err
	}

	returnStakingIns.SetShardID(byte(shardID))
	returnStakingIns, err = returnStakingIns.SetStakingTXIDs(strings.Split(instruction[3], SPLITTER))
	if err != nil {
		return nil, err
	}

	percentRetunrsStr := strings.Split(instruction[4], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		percentReturns[i] = uint(tempPercent)
	}
	returnStakingIns.SetPercentReturns(percentReturns)
	return returnStakingIns, err
}

func ValidateReturnStakingInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RETURN_ACTION {
		return fmt.Errorf("invalid return staking action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	txStakes := strings.Split(instruction[3], SPLITTER)
	for _, txStake := range txStakes {
		_, err := common.Hash{}.NewHashFromStr(txStake)
		if err != nil {
			return fmt.Errorf("invalid tx stake %+v", err)
		}
	}
	percentRetunrsStr := strings.Split(instruction[4], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid percent return %+v", err)
		}
		percentReturns[i] = uint(tempPercent)
	}
	if len(publicKeys) != len(txStakes) {
		return fmt.Errorf("invalid public key & tx stake length, %+v", instruction)
	}
	if len(percentReturns) != len(txStakes) {
		return fmt.Errorf("invalid reward percentReturns & tx stake length, %+v", instruction)
	}
	if len(percentReturns) != len(publicKeys) {
		return fmt.Errorf("invalid reward percentReturns & publicKeys length, %+v", instruction)
	}
	return nil
}
