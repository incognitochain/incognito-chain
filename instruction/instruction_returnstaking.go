package instruction

import "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

type ReturnStakeIns struct {
	PublicKey     string
	ShardID       byte
	StakingTXID   string
	PercentReturn uint
}

func NewReturnStakeInsWithValue(
	publicKey string,
	sID byte,
	txStake string,
	pReturn uint,
) *ReturnStakeIns {
	return &ReturnStakeIns{
		PublicKey:     publicKey,
		ShardID:       sID,
		StakingTXID:   txStake,
		PercentReturn: pReturn,
	}
}

func NewReturnStakeIns() *ReturnStakeIns {
	return &ReturnStakeIns{}
}

func (rsI *ReturnStakeIns) SetShardID(sID byte) error {
	rsI.ShardID = sID
	return nil
}

func (rsI *ReturnStakeIns) SetStakingTXID(txID string) error {
	rsI.StakingTXID = txID
	return nil
}

func (rsI *ReturnStakeIns) SetPercentReturn(percent uint) error {
	rsI.PercentReturn = percent
	return nil
}

func (rsI *ReturnStakeIns) GetType() string {
	return RETURN_ACTION
}

func (rsI *ReturnStakeIns) GetShardID() byte {
	return rsI.ShardID
}

func (rsI *ReturnStakeIns) GetPercentReturn() uint {
	return rsI.PercentReturn
}

func (rsI *ReturnStakeIns) GetStakingTX() string {
	return rsI.StakingTXID
}

func (rsI *ReturnStakeIns) GetPublicKey() string {
	return rsI.PublicKey
}

func (rsI *ReturnStakeIns) ToString() []string {
	// returnStakeInsStr := []string{STAKE_ACTION}
	// returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.PublicKeys, SPLITTER))
	// returnStakeInsStr = append(returnStakeInsStr, rsI.Chain)
	// returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.TxStakes, SPLITTER))
	// returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.RewardReceivers, SPLITTER))
	// tempStopAutoStakeFlag := []string{}
	// for _, v := range rsI.AutoStakingFlag {
	// 	if v == true {
	// 		tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, TRUE)
	// 	} else {
	// 		tempStopAutoStakeFlag = append(tempStopAutoStakeFlag, FALSE)
	// 	}
	// }
	// stakeInstructionStr = append(stakeInstructionStr, strings.Join(tempStopAutoStakeFlag, SPLITTER))
	// return stakeInstructionStr
	return []string{}
}

func (rsI *ReturnStakeIns) InsertIntoStateDB(sDB *statedb.StateDB) error {
	return nil
}
