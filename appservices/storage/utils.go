package storage

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
)

func StoreLatestBeaconFinalState(ctx context.Context, beacon *data.Beacon) error {
	Logger.log.Infof("Store beacon with block hash %v and block height %d", beacon.BlockHash, beacon.Height)
	beaconState := getBeaconFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetBeaconStorer().StoreBeaconState(ctx, beaconState); err != nil {
		return err
	}
	Logger.log.Infof("This beacon contain %d PDE Share ", len(beacon.PDEShare))
	if len(beacon.PDEShare) > 0 {
		pdeShares := getPDEShareFromBeaconState(beacon)
		for _, pdeShare := range pdeShares {
			GetDBDriver(MONGODB).GetPDEShareStorer().StorePDEShare(ctx, pdeShare)
		}
	}
	return nil
}

func getBeaconFromBeaconState(beacon *data.Beacon) model.BeaconState {
	return model.BeaconState{
		ShardID:                                beacon.ShardID,
		BlockHash:                              beacon.BlockHash,
		PreviousBlockHash:                      beacon.PreviousBlockHash,
		BestShardHash:                          beacon.BestShardHash,
		BestShardHeight:                        beacon.BestShardHeight,
		Epoch:                                  beacon.Epoch,
		Height:                                 beacon.Height,
		ProposerIndex:                          beacon.ProposerIndex,
		BeaconCommittee:                        beacon.BeaconCommittee,
		BeaconPendingValidator:                 beacon.BeaconPendingValidator,
		CandidateBeaconWaitingForCurrentRandom: beacon.CandidateBeaconWaitingForNextRandom,
		CandidateShardWaitingForCurrentRandom:  beacon.CandidateShardWaitingForCurrentRandom,
		CandidateBeaconWaitingForNextRandom:    beacon.CandidateBeaconWaitingForNextRandom,
		CandidateShardWaitingForNextRandom:     beacon.CandidateShardWaitingForNextRandom,
		ShardCommittee:                         beacon.ShardCommittee,
		ShardPendingValidator:                  beacon.ShardPendingValidator,
		AutoStaking:                            beacon.AutoStaking,
		CurrentRandomNumber:                    beacon.CurrentRandomNumber,
		CurrentRandomTimeStamp:                 beacon.CurrentRandomTimeStamp,
		MaxBeaconCommitteeSize:                 beacon.MaxBeaconCommitteeSize,
		MinBeaconCommitteeSize:                 beacon.MinBeaconCommitteeSize,
		MaxShardCommitteeSize:                  beacon.MaxShardCommitteeSize,
		MinShardCommitteeSize:                  beacon.MinShardCommitteeSize,
		ActiveShards:                           beacon.ActiveShards,
		LastCrossShardState:                    beacon.LastCrossShardState,
		Time:                                   beacon.Time,
		ConsensusAlgorithm:                     beacon.ConsensusAlgorithm,
		ShardConsensusAlgorithm:                beacon.ShardConsensusAlgorithm,
		Instruction:                            beacon.Instruction,
	}
}

func getPDEShareFromBeaconState(beacon *data.Beacon) []model.PDEShare {
	pdeShares := make([]model.PDEShare, 0, len(beacon.PDEShare))
	for _, share := range beacon.PDEShare {
		pdeShares = append(pdeShares, model.PDEShare{
			BeaconBlockHash:    beacon.BlockHash,
			BeaconEpoch:        beacon.Epoch,
			BeaconHeight:       beacon.Height,
			BeaconTime:         beacon.Time,
			Token1ID:           share.Token1ID,
			Token2ID:           share.Token2ID,
			ContributorAddress: share.ContributorAddress,
			Amount:             share.Amount,
		})
	}
	return pdeShares
}

func StoreLatestShardFinalState(ctx context.Context, shard *data.Shard) error {
	Logger.log.Infof("Store shard with block hash %v and block height %d of Shard ID", shard.BlockHash, shard.Height, shard.ShardID)
	shardState := getShardFromShardState(shard)
	if err := GetDBDriver(MONGODB).GetShardStorer().StoreShardState(ctx, shardState); err != nil {
		return err
	}
	if len(shard.Transactions) > 0 {
		transactions := getTransactionFromShardState(shard)
		Logger.log.Infof("Store number of transactions %d", len(transactions))
		for _, transaction := range transactions {
			GetDBDriver(MONGODB).GetTransactionStorer().StoreTransaction(ctx, transaction)
		}

		outputCoins := getOutputCoinFromShardState(shard)
		inputCoins := getInputCoinFromShardState(shard)

		for _, inputCoin := range inputCoins {
			GetDBDriver(MONGODB).GetInputCoinStorer().StoreInputCoin(ctx, inputCoin)
		}

		for _, outputCoin := range outputCoins {
			GetDBDriver(MONGODB).GetOutputCoinStorer().StoreOutputCoin(ctx, outputCoin)
		}

	}

	if len(shard.Commitments) > 0 {
		commitments := getCommitmentFromShardState(shard)
		Logger.log.Infof("Store commitment with size %d", len(commitments))

		for _, commitment := range commitments {
			//Logger.log.Infof("Store commitment %v", commitment)
			GetDBDriver(MONGODB).GetCommitmentStorer().StoreCommitment(ctx, commitment)
		}
	}
	return nil
}

func getShardFromShardState(shard *data.Shard) model.ShardState {
	return model.ShardState{
		ShardID:                shard.ShardID,
		BlockHash:              shard.BlockHash,
		PreviousBlockHash:      shard.PreviousBlockHash,
		Height:                 shard.Height,
		Version:                shard.Version,
		TxRoot:                 shard.TxRoot,
		Time:                   shard.Time,
		TxHashes:               shard.TxHashes,
		Txs:                    shard.Txs,
		BlockProducer:          shard.BlockProducer,
		ValidationData:         shard.ValidationData,
		ConsensusType:          shard.ConsensusType,
		Data:                   shard.Data,
		BeaconHeight:           shard.BeaconHeight,
		BeaconBlockHash:        shard.BeaconBlockHash,
		Round:                  shard.Round,
		Epoch:                  shard.Epoch,
		Reward:                 shard.Reward,
		RewardBeacon:           shard.RewardBeacon,
		Fee:                    shard.Fee,
		Size:                   shard.Size,
		Instruction:            shard.Instruction,
		CrossShardBitMap:       shard.CrossShardBitMap,
		NumTxns:                shard.NumTxns,
		TotalTxns:              shard.TotalTxns,
		NumTxnsExcludeSalary:   shard.NumTxnsExcludeSalary,
		TotalTxnsExcludeSalary: shard.TotalTxnsExcludeSalary,
		ActiveShards:           shard.ActiveShards,
		ConsensusAlgorithm:     shard.ConsensusType,
		NumOfBlocksByProducers: shard.NumOfBlocksByProducers,
	}
}

func getTransactionFromShardState(shard *data.Shard) []model.Transaction {
	transactions := make([]model.Transaction, 0, len(shard.Transactions))
	for _, transaction := range shard.Transactions {
		transactions = append(transactions, model.Transaction{
			ShardId:    shard.ShardID,
			ShardHash: shard.BlockHash,
			ShardHeight: shard.BeaconHeight,
			Hash:      transaction.Hash,
			Version:   transaction.Version,
			Type:      transaction.Type,
			LockTime:  transaction.LockTime,
			Fee:       transaction.Fee,
			Info:      transaction.Info,
			SigPubKey: transaction.SigPubKey,
			Sig:       transaction.Sig,
			Proof:     transaction.Proof,
			Metadata:  transaction.Metadata,
			PubKeyLastByteSender: transaction.PubKeyLastByteSender,
		})
	}
	return transactions
}


func getInputCoinFromShardState(shard *data.Shard) []model.InputCoin {
	inputCoins := make([]model.InputCoin, 0, len(shard.Transactions))
	for _, transaction := range shard.Transactions {
		for _, input := range transaction.InputCoins {
			inputCoin := model.InputCoin{
				ShardId:         shard.ShardID,
				ShardHash:       shard.BlockHash,
				ShardHeight:     shard.BeaconHeight,
				TransactionHash: transaction.Hash,
				Value:           input.CoinDetails.Value,
				Info:            input.CoinDetails.Info,
			}
			if input.CoinDetails.PublicKey != nil {
				inputCoin.PublicKey = input.CoinDetails.PublicKey.ToBytesS()
			}
			if input.CoinDetails.CoinCommitment != nil {
				inputCoin.CoinCommitment = input.CoinDetails.CoinCommitment.ToBytesS()
			}
			if input.CoinDetails.SNDerivator != nil {
				inputCoin.SNDerivator = input.CoinDetails.SNDerivator.ToBytesS()
			}
			if input.CoinDetails.SerialNumber != nil {
				inputCoin.SerialNumber = input.CoinDetails.SerialNumber.ToBytesS()
			}
			if input.CoinDetails.Randomness != nil {
				inputCoin.Randomness = input.CoinDetails.Randomness.ToBytesS()
			}
			inputCoins = append(inputCoins, inputCoin)
		}

	}
	return inputCoins
}

func getOutputCoinFromShardState(shard *data.Shard) []model.OutputCoin {
	outputCoins := make([]model.OutputCoin, 0, len(shard.Transactions))
	for _, transaction := range shard.Transactions {
		for _, output := range transaction.OutputCoins {
			outputCoin := model.OutputCoin{
				ShardId:         shard.ShardID,
				ShardHash:       shard.BlockHash,
				ShardHeight:     shard.BeaconHeight,
				TransactionHash: transaction.Hash,
				Value:           output.CoinDetails.Value,
				Info:            output.CoinDetails.Info,
			}
			if output.CoinDetails.PublicKey != nil {
				outputCoin.PublicKey = output.CoinDetails.PublicKey.ToBytesS()
			}
			if output.CoinDetails.CoinCommitment != nil {
				outputCoin.CoinCommitment = output.CoinDetails.CoinCommitment.ToBytesS()
			}
			if output.CoinDetails.SNDerivator != nil {
				outputCoin.SNDerivator = output.CoinDetails.SNDerivator.ToBytesS()
			}
			if output.CoinDetails.SerialNumber != nil {
				outputCoin.SerialNumber = output.CoinDetails.SerialNumber.ToBytesS()
			}
			if output.CoinDetails.Randomness != nil {
				outputCoin.Randomness = output.CoinDetails.Randomness.ToBytesS()
			}
			outputCoins = append(outputCoins, outputCoin)
		}

	}
	return outputCoins
}

func getCommitmentFromShardState(shard *data.Shard) []model.Commitment {
	commitments := make([]model.Commitment, 0, len(shard.Commitments))

	for _, commitment:= range shard.Commitments {
		commitments = append(commitments, model.Commitment{
			ShardHash:       shard.BlockHash,
			ShardHeight:     shard.Height,
			TransactionHash: commitment.TransactionHash,
			TokenID:         commitment.TokenID,
			ShardId:         commitment.ShardID,
			Commitment:      commitment.Commitment,
			Index:           commitment.Index,
		})
	}
	return commitments
}