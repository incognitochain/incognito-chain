package instruction

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ReturnBeaconStakeInstruction struct {
	PublicKeys       []string
	PublicKeysStruct []incognitokey.CommitteePublicKey
	StakingTXIDs     [][]string
	StakingTxHashes  [][]common.Hash
	PercentReturns   []uint
	Reasons          []byte
}

func NewReturnBeaconStakeInsWithValue(
	publicKeys []string,
	txStake [][]string,
	reason []byte,
) *ReturnBeaconStakeInstruction {
	rsI := &ReturnBeaconStakeInstruction{}
	rsI, _ = rsI.SetPublicKeys(publicKeys)
	rsI, _ = rsI.SetStakingTXIDs(txStake)
	for _, _ = range publicKeys {
		rsI.PercentReturns = append(rsI.PercentReturns, 100)
	}
	if len(reason) != 0 {
		rsI.SetReasons(reason)
	}
	return rsI
}

func NewReturnBeaconStakeIns() *ReturnBeaconStakeInstruction {
	return &ReturnBeaconStakeInstruction{}
}

func (rsI *ReturnBeaconStakeInstruction) IsEmpty() bool {
	return reflect.DeepEqual(rsI, NewReturnStakeIns()) ||
		len(rsI.PublicKeysStruct) == 0 && len(rsI.PublicKeys) == 0
}

func (rsI *ReturnBeaconStakeInstruction) SetPublicKeys(publicKeys []string) (*ReturnBeaconStakeInstruction, error) {
	if publicKeys == nil {
		return nil, errors.New("Public Keys Are Null")
	}
	rsI.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	rsI.PublicKeysStruct = publicKeyStructs
	return rsI, nil
}

func (rsI *ReturnBeaconStakeInstruction) SetStakingTXIDs(txIDs [][]string) (*ReturnBeaconStakeInstruction, error) {
	if txIDs == nil {
		return nil, errors.New("Tx Hashes Are Null")
	}
	rsI.StakingTXIDs = txIDs
	rsI.StakingTxHashes = make([][]common.Hash, len(txIDs))
	for i, stakingTxIDs := range rsI.StakingTXIDs {
		rsI.StakingTxHashes[i] = make([]common.Hash, len(txIDs[i]))
		for j, v := range stakingTxIDs {
			temp, err := common.Hash{}.NewHashFromStr(v)
			if err != nil {
				return rsI, err
			}
			rsI.StakingTxHashes[i][j] = *temp
		}
	}
	return rsI, nil
}

func (rsI *ReturnBeaconStakeInstruction) SetPercentReturns(percentReturns []uint) error {
	rsI.PercentReturns = percentReturns
	return nil
}

func (rsI *ReturnBeaconStakeInstruction) SetReasons(reason []byte) error {
	rsI.Reasons = reason
	return nil
}

func (rsI *ReturnBeaconStakeInstruction) GetType() string {
	return RETURN_BEACON_ACTION
}

func (rsI *ReturnBeaconStakeInstruction) GetPercentReturns() []uint {
	return rsI.PercentReturns
}

func (rsI *ReturnBeaconStakeInstruction) GetStakingTX() [][]string {
	return rsI.StakingTXIDs
}

func (rsI *ReturnBeaconStakeInstruction) GetPublicKey() []string {
	return rsI.PublicKeys
}

func (rsI *ReturnBeaconStakeInstruction) GetReason() []byte {
	return rsI.Reasons
}

func (rsI *ReturnBeaconStakeInstruction) ToString() []string {
	returnStakeInsStr := []string{RETURN_BEACON_ACTION}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.PublicKeys, SPLITTER))
	txIDsPerPK := []string{}
	for _, txIDs := range rsI.StakingTXIDs {
		txIDsPerPK = append(txIDsPerPK, strings.Join(txIDs, "-"))
	}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(txIDsPerPK, SPLITTER))
	percentReturnsStr := make([]string, len(rsI.PercentReturns))
	for i, v := range rsI.PercentReturns {
		percentReturnsStr[i] = strconv.Itoa(int(v))
	}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(percentReturnsStr, SPLITTER))
	if len(rsI.Reasons) != 0 {
		reasonStrs := make([]string, len(rsI.Reasons))
		for i, v := range rsI.Reasons {
			reasonStrs[i] = strconv.Itoa(int(v))
		}
		returnStakeInsStr = append(returnStakeInsStr, strings.Join(reasonStrs, SPLITTER))

	}
	return returnStakeInsStr
}

func (rsI *ReturnBeaconStakeInstruction) AddNewRequest(publicKey string, stakingTxs []string) {
	rsI.PublicKeys = append(rsI.PublicKeys, publicKey)
	publicKeyStruct, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{publicKey})
	rsI.PublicKeysStruct = append(rsI.PublicKeysStruct, publicKeyStruct[0])
	rsI.StakingTXIDs = append(rsI.StakingTXIDs, stakingTxs)
	stakingTxHashes := []common.Hash{}
	for _, stakingTx := range stakingTxs {
		stakingTxHash, _ := common.Hash{}.NewHashFromStr(stakingTx)
		stakingTxHashes = append(stakingTxHashes, *stakingTxHash)
	}
	rsI.StakingTxHashes = append(rsI.StakingTxHashes, stakingTxHashes)
	rsI.PercentReturns = append(rsI.PercentReturns, 100)
}

func ValidateAndImportReturnBeaconStakingInstructionFromString(instruction []string) (*ReturnBeaconStakeInstruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}

func BuildReturnBeaconStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}

func ImportReturnBeaconStakingInstructionFromString(instruction []string) (*ReturnBeaconStakeInstruction, error) {
	returnStakingIns := NewReturnBeaconStakeIns()
	var err error
	returnStakingIns, err = returnStakingIns.SetPublicKeys(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}
	stakingTxsRaw := strings.Split(instruction[2], SPLITTER)
	stakingTxIDs := make([][]string, len(returnStakingIns.PublicKeys))
	for i, _ := range stakingTxIDs {
		stakingTxIDsPerPK := strings.Split(stakingTxsRaw[i], "-")
		stakingTxIDs[i] = make([]string, len(stakingTxIDsPerPK))
		for j, txID := range stakingTxIDsPerPK {
			stakingTxIDs[i][j] = txID
		}
	}

	returnStakingIns, err = returnStakingIns.SetStakingTXIDs(stakingTxIDs)
	if err != nil {
		return nil, err
	}

	percentRetunrsStr := strings.Split(instruction[3], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		percentReturns[i] = uint(tempPercent)
	}
	returnStakingIns.SetPercentReturns(percentReturns)
	if len(instruction) == 5 {
		reasonsStr := strings.Split(instruction[4], SPLITTER)
		reasons := make([]byte, len(reasonsStr))
		for i, v := range reasonsStr {
			reason, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			reasons[i] = byte(reason)
		}
		returnStakingIns.SetReasons(reasons)
	}
	return returnStakingIns, err
}

func ValidateReturnBeaconStakingInstructionSanity(instruction []string) error {
	if (len(instruction) != 4) && (len(instruction) != 5) {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RETURN_BEACON_ACTION {
		return fmt.Errorf("invalid return staking action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return err
	}

	stakingTxsRaws := strings.Split(instruction[2], SPLITTER)
	for _, stakingTxsRaw := range stakingTxsRaws {
		stakingTxIDsPerPK := strings.Split(stakingTxsRaw, "-")
		for _, txID := range stakingTxIDsPerPK {
			_, err := common.Hash{}.NewHashFromStr(txID)
			if err != nil {
				log.Println("err:", err)
				return fmt.Errorf("invalid tx return staking %+v", err)
			}
		}
	}
	percentRetunrsStr := strings.Split(instruction[3], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid percent return %+v", err)
		}
		percentReturns[i] = uint(tempPercent)
	}
	if len(publicKeys) != len(stakingTxsRaws) {
		return fmt.Errorf("invalid public key & tx staking txs length, %+v", instruction)
	}
	if len(percentReturns) != len(stakingTxsRaws) {
		return fmt.Errorf("invalid reward percentReturns & tx stakings length, %+v", instruction)
	}
	if len(percentReturns) != len(publicKeys) {
		return fmt.Errorf("invalid reward percentReturns & public Keys length, %+v", instruction)
	}
	return nil
}
