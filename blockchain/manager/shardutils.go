package manager

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

func returnStakingFromIns(
	insStake instruction.ReturnStakeIns,
	producerPrivateKey *privacy.PrivateKey,
	incDB incdb.Database,
	txStateDB *statedb.StateDB,
) (
	metadata.Transaction,
	error,
) {
	stakeTxHash := &common.Hash{}
	stakeTxHash, _ = stakeTxHash.NewHashFromStr(insStake.StakingTXID)
	txStake, err := GetTxDataByHash(incDB, *stakeTxHash)
	if err != nil {
		//TODO find the correctly way to handle error here, panic or continue or do something else?
		return nil, err
	}
	returnStakingMeta := metadata.NewReturnStakingMetaFromStakingTx(
		txStake,
	)
	stakeAmount := txStake.CalculateTxValue()
	returnStakingTx := new(transaction.Tx)
	err = returnStakingTx.InitTxSalary(
		stakeAmount*uint64(insStake.PercentReturn)/100,
		&returnStakingMeta.StakerAddress,
		producerPrivateKey,
		txStateDB,
		returnStakingMeta,
	)
	if err != nil {
		return nil, err
	}
	return returnStakingTx, nil
}

func stakeInsFromTx(
	shardID byte,
	meta metadata.Metadata,
	txHash string,
	mapInsStake map[string]*instruction.StakeInstruction,
) error {
	stakeMeta, ok := meta.(*metadata.StakingMetadata)
	if !ok {
		return errors.Errorf("Can not parse this metadata %v", meta.Hash())
	}
	chainName := "beacon"
	if stakeMeta.GetType() == metadata.ShardStakingMeta {
		chainName = fmt.Sprintf("shard-%v", shardID)
	}
	insStake, ok := mapInsStake[chainName]
	if !ok {
		insStake = instruction.NewStakeInstruction()
		insStake.Chain = chainName
	}
	insStake.PublicKeys = append(insStake.PublicKeys, stakeMeta.CommitteePublicKey)
	insStake.TxStakes = append(insStake.TxStakes, txHash)
	insStake.RewardReceivers = append(insStake.RewardReceivers, stakeMeta.RewardReceiverPaymentAddress)
	insStake.AutoStakingFlag = append(insStake.AutoStakingFlag, stakeMeta.AutoReStaking)
	mapInsStake[chainName] = insStake
	return nil
}

func stopInsFromTx(
	shardID byte,
	meta metadata.Metadata,
	insStop *instruction.StopAutoStakeInstruction,
) error {
	stopAutoStakingMetadata, ok := meta.(*metadata.StopAutoStakingMetadata)
	if !ok {
		return errors.Errorf("Can not parse this metadata %v", meta.Hash())
	}
	insStop.PublicKeys = append(insStop.PublicKeys, stopAutoStakingMetadata.CommitteePublicKey)
	return nil
}

//unstakeInsFromTx : Build unstake instruction
// from unstake tx of ShardBlock
// for making BeaconBlock from below instruction
func unstakeInsFromTx(
	shardID byte,
	meta metadata.Metadata,
) (*instruction.UnstakeInstruction, error) {
	unstakingMetadata, ok := meta.(*metadata.UnStakingMetadata)
	if !ok {
		return nil, errors.Errorf("Can not parse this metadata %v", meta.Hash())
	}

	unstakingInstruction, err := instruction.
		ValidateAndImportUnstakeInstructionFromString(
			[]string{instruction.UNSTAKE_ACTION, unstakingMetadata.CommitteePublicKey})

	return unstakingInstruction, err
}
