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
		panic(err)
		return err
	}
	//PDE
	Logger.log.Infof("This beacon contain %d PDE Share ", len(beacon.PDEShare))
	if len(beacon.PDEShare) > 0 {
		pdeShares := getPDEShareFromBeaconState(beacon)
		for _, pdeShare := range pdeShares {
			if err := GetDBDriver(MONGODB).GetPDEShareStorer().StorePDEShare(ctx, pdeShare); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.PDEPoolPair) > 0 {
		pdePoolPairs := getPDEPoolForPairStateFromBeaconState(beacon)
		for _, pdePoolPair := range pdePoolPairs {
			if err := GetDBDriver(MONGODB).GetPDEPoolForPairStateStorer().StorePDEPoolForPairState(ctx, pdePoolPair); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.PDETradingFee) > 0 {
		pdeTradingFees := getPDETradingFeeFromBeaconState(beacon)
		for _, pdeTradingFee := range pdeTradingFees {
			if err := GetDBDriver(MONGODB).GetPDETradingFeeStorer().StorePDETradingFee(ctx, pdeTradingFee); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.WaitingPDEContributionState) > 0 {
		waitingPDEContributionStates := getWaitingPDEContributionStateFromBeaconState(beacon)
		for _, waiting := range waitingPDEContributionStates {
			if err := GetDBDriver(MONGODB).GetWaitingPDEContributionStorer().StoreWaitingPDEContribution(ctx, waiting); err != nil {
				panic(err)
				return err
			}
		}
	}

	//Portal
	if len(beacon.Custodian) > 0 {
		custodians := getCustodianFromBeaconState(beacon)
		for _, custodian := range custodians {
			if err := GetDBDriver(MONGODB).GetCustodianStorer().StoreCustodian(ctx, custodian); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.WaitingPortingRequest) > 0 {
		waitingPortingRequests := getWaitingPortingRequestFromBeaconState(beacon)
		for _, waiting := range waitingPortingRequests {
			if err := GetDBDriver(MONGODB).GetWaitingPortingRequestStorer().StoreWaitingPortingRequest(ctx, waiting); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.MatchedRedeemRequest) > 0 {
		matchedRedeemRequests := getMatchedRedeemRequestFromBeaconState(beacon)
		for _, matched := range matchedRedeemRequests {
			if err := GetDBDriver(MONGODB).GetMatchedRedeemRequestStorer().StoreMatchedRedeemRequest(ctx, matched); err != nil {
				panic(err)
				return err
			}
		}
	}

	if len(beacon.WaitingRedeemRequest) > 0 {
		waitingRedeemRequests := getWaitingRedeemRequestFromBeaconState(beacon)
		for _, waiting := range waitingRedeemRequests {
			if err := GetDBDriver(MONGODB).GetWaitingRedeemRequestStorer().StoreWaitingRedeemRequest(ctx, waiting); err != nil {
				panic(err)
				return err
			}
		}
	}

	finalExchangeRates := getFinalExchangeRatesFromBeaconState(beacon)
	for _, final := range finalExchangeRates {
		if err := GetDBDriver(MONGODB).GetFinalExchangeRatesStorer().StoreFinalExchangeRates(ctx, final); err != nil {
			panic(err)
			return err
		}
	}

	lockedCollaterals := getLockedCollateralFromBeaconState(beacon)
	for _, locked := range lockedCollaterals {
		if err := GetDBDriver(MONGODB).GetLockedCollateralStorer().StoreLockedCollateral(ctx, locked); err != nil {
			panic(err)
			return err
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
func getWaitingPDEContributionStateFromBeaconState(beacon *data.Beacon) []model.WaitingPDEContribution {
	waitingPDEContributions := make([]model.WaitingPDEContribution, 0, len(beacon.WaitingPDEContributionState))
	for _, waiting := range beacon.WaitingPDEContributionState {
		waitingPDEContributions = append(waitingPDEContributions, model.WaitingPDEContribution{
			BeaconBlockHash:    beacon.BlockHash,
			BeaconEpoch:        beacon.Epoch,
			BeaconHeight:       beacon.Height,
			BeaconTime:         beacon.Time,
			PairID:             waiting.PairID,
			ContributorAddress: waiting.ContributorAddress,
			TokenID:            waiting.TokenID,
			Amount:             waiting.Amount,
			TXReqID:            waiting.TXReqID,
		})
	}
	return waitingPDEContributions
}

func getPDETradingFeeFromBeaconState(beacon *data.Beacon) []model.PDETradingFee {
	pdeTradingFees := make([]model.PDETradingFee, 0, len(beacon.PDETradingFee))
	for _, pdeTradingFee := range beacon.PDETradingFee {
		pdeTradingFees = append(pdeTradingFees, model.PDETradingFee{
			BeaconBlockHash:    beacon.BlockHash,
			BeaconEpoch:        beacon.Epoch,
			BeaconHeight:       beacon.Height,
			BeaconTime:         beacon.Time,
			Token1ID:           pdeTradingFee.Token1ID,
			Token2ID:           pdeTradingFee.Token2ID,
			ContributorAddress: pdeTradingFee.ContributorAddress,
			Amount:             pdeTradingFee.Amount,
		})
	}
	return pdeTradingFees
}

func getPDEPoolForPairStateFromBeaconState(beacon *data.Beacon) []model.PDEPoolForPair {
	pdeFoolForPairs := make([]model.PDEPoolForPair, 0, len(beacon.PDEPoolPair))
	for _, pdeFoolForPair := range beacon.PDEPoolPair {
		pdeFoolForPairs = append(pdeFoolForPairs, model.PDEPoolForPair{
			BeaconBlockHash: beacon.BlockHash,
			BeaconEpoch:     beacon.Epoch,
			BeaconHeight:    beacon.Height,
			BeaconTime:      beacon.Time,
			Token1ID:        pdeFoolForPair.Token1ID,
			Token1PoolValue: pdeFoolForPair.Token1PoolValue,
			Token2ID:        pdeFoolForPair.Token2ID,
			Token2PoolValue: pdeFoolForPair.Token2PoolValue,
		})
	}
	return pdeFoolForPairs
}

func getCustodianFromBeaconState(beacon *data.Beacon) []model.Custodian {
	custodians := make([]model.Custodian, 0, len(beacon.Custodian))
	for _, custodian := range beacon.Custodian {
		custodians = append(custodians, model.Custodian{
			BeaconBlockHash:        beacon.BlockHash,
			BeaconEpoch:            beacon.Epoch,
			BeaconHeight:           beacon.Height,
			BeaconTime:             beacon.Time,
			IncognitoAddress:       custodian.IncognitoAddress,
			TotalCollateral:        custodian.TotalCollateral,
			FreeCollateral:         custodian.FreeCollateral,
			HoldingPubTokens:       custodian.HoldingPubTokens,
			LockedAmountCollateral: custodian.LockedAmountCollateral,
			RemoteAddresses:        custodian.RemoteAddresses,
			RewardAmount:           custodian.RewardAmount,
		})
	}
	return custodians
}

func getWaitingPortingRequestFromBeaconState(beacon *data.Beacon) []model.WaitingPortingRequest {
	waitingPortingRequests := make([]model.WaitingPortingRequest, 0, len(beacon.WaitingPortingRequest))
	for _, w := range beacon.WaitingPortingRequest {
		waitingPortingRequests = append(waitingPortingRequests, model.WaitingPortingRequest{
			BeaconBlockHash:     beacon.BlockHash,
			BeaconEpoch:         beacon.Epoch,
			BeaconHeight:        beacon.Height,
			BeaconTime:          beacon.Time,
			UniquePortingID:     w.UniquePortingID,
			TokenID:             w.TokenID,
			PorterAddress:       w.PorterAddress,
			Amount:              w.Amount,
			Custodians:          w.Custodians,
			PortingFee:          w.PortingFee,
			WaitingBeaconHeight: w.BeaconHeight,
			TXReqID:             w.TXReqID,
		})
	}
	return waitingPortingRequests
}

func getFinalExchangeRatesFromBeaconState(beacon *data.Beacon) []model.FinalExchangeRate {
	finalExchangeRates := make([]model.FinalExchangeRate, 0, len(beacon.FinalExchangeRates.Rates))
	for key, amount := range beacon.FinalExchangeRates.Rates {
		finalExchangeRates = append(finalExchangeRates, model.FinalExchangeRate{
			BeaconBlockHash: beacon.BlockHash,
			BeaconEpoch:     beacon.Epoch,
			BeaconHeight:    beacon.Height,
			BeaconTime:      beacon.Time,
			Amount:          amount.Amount,
			TokenID:         key,
		})
	}
	return finalExchangeRates
}

func getMatchedRedeemRequestFromBeaconState(beacon *data.Beacon) []model.RedeemRequest {
	redeemRequests := make([]model.RedeemRequest, 0, len(beacon.MatchedRedeemRequest))
	for _, matchedRedeem := range beacon.MatchedRedeemRequest {
		redeemRequests = append(redeemRequests, model.RedeemRequest{
			BeaconBlockHash:       beacon.BlockHash,
			BeaconEpoch:           beacon.Epoch,
			BeaconHeight:          beacon.Height,
			BeaconTime:            beacon.Time,
			UniqueRedeemID:        matchedRedeem.UniqueRedeemID,
			TokenID:               matchedRedeem.TokenID,
			RedeemerAddress:       matchedRedeem.RedeemerAddress,
			RedeemerRemoteAddress: matchedRedeem.RedeemerRemoteAddress,
			RedeemAmount:          matchedRedeem.RedeemAmount,
			Custodians:            matchedRedeem.Custodians,
			RedeemFee:             matchedRedeem.RedeemFee,
			RedeemBeaconHeight:    matchedRedeem.BeaconHeight,
			TXReqID:               matchedRedeem.TXReqID,
		})
	}
	return redeemRequests
}

func getWaitingRedeemRequestFromBeaconState(beacon *data.Beacon) []model.RedeemRequest {
	redeemRequests := make([]model.RedeemRequest, 0, len(beacon.WaitingRedeemRequest))
	for _, waitingRedeem := range beacon.WaitingRedeemRequest {
		redeemRequests = append(redeemRequests, model.RedeemRequest{
			BeaconBlockHash:       beacon.BlockHash,
			BeaconEpoch:           beacon.Epoch,
			BeaconHeight:          beacon.Height,
			BeaconTime:            beacon.Time,
			UniqueRedeemID:        waitingRedeem.UniqueRedeemID,
			TokenID:               waitingRedeem.TokenID,
			RedeemerAddress:       waitingRedeem.RedeemerAddress,
			RedeemerRemoteAddress: waitingRedeem.RedeemerRemoteAddress,
			RedeemAmount:          waitingRedeem.RedeemAmount,
			Custodians:            waitingRedeem.Custodians,
			RedeemFee:             waitingRedeem.RedeemFee,
			RedeemBeaconHeight:    waitingRedeem.BeaconHeight,
			TXReqID:               waitingRedeem.TXReqID,
		})
	}
	return redeemRequests
}

func getLockedCollateralFromBeaconState(beacon *data.Beacon) []model.LockedCollateral {
	lockedCollaterals := make([]model.LockedCollateral, 0, len(beacon.LockedCollateralState.LockedCollateralDetail))
	for key, lockedDetail := range beacon.LockedCollateralState.LockedCollateralDetail {
		lockedCollaterals = append(lockedCollaterals, model.LockedCollateral{
			BeaconBlockHash:                 beacon.BlockHash,
			BeaconEpoch:                     beacon.Epoch,
			BeaconHeight:                    beacon.Height,
			BeaconTime:                      beacon.Time,
			TotalLockedCollateralForRewards: beacon.LockedCollateralState.TotalLockedCollateralForRewards,
			CustodianAddress:                key,
			Amount:                          lockedDetail,
		})
	}
	return lockedCollaterals
}

func StoreLatestShardFinalState(ctx context.Context, shard *data.Shard) error {
	Logger.log.Infof("Store shard with block hash %v and block height %d of Shard ID %d", shard.BlockHash, shard.Height, shard.ShardID)
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

		inputCoins := getInputCoinFromShardState(shard)

		for _, inputCoin := range inputCoins {
			GetDBDriver(MONGODB).GetInputCoinStorer().StoreInputCoin(ctx, inputCoin)
		}
	}

	if len(shard.OutputCoins) > 0 {
		outputCoins := getOutputCoinForThisShardFromShardState(shard)
		for _, outputCoin := range outputCoins {
			GetDBDriver(MONGODB).GetOutputCoinStorer().StoreOutputCoin(ctx, outputCoin)
		}
	}

	if len(shard.OutputCoins) > 0 {
		outputCoins := getCrossShardOutputCoinFromShardState(shard)
		Logger.log.Debugf("Store cross shard output coin with size %d", len(outputCoins))
		for _, outputCoin := range outputCoins {
			GetDBDriver(MONGODB).GetCrossShardOutputCoinStorer().StoreCrossShardOutputCoin(ctx, outputCoin)
		}
	}

	if len(shard.Commitments) > 0 {
		commitments := getCommitmentFromShardState(shard)
		Logger.log.Infof("Store commitment with size %d", len(commitments))
		for _, commitment := range commitments {
			Logger.log.Debugf("Store commitment %v", commitment)
			GetDBDriver(MONGODB).GetCommitmentStorer().StoreCommitment(ctx, commitment)
		}
	}

	if len(shard.TokenState) > 0 {
		tokenState := GetTokenStateFromShardState(shard)
		if err := GetDBDriver(MONGODB).GetTokenStateStorer().StoreTokenState(ctx, tokenState); err != nil {
			panic(err)
			return err
		}

	}
	//fmt.Scanln()
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
		newTransaction := model.Transaction{
			ShardId:              shard.ShardID,
			ShardHash:            shard.BlockHash,
			ShardHeight:          shard.BeaconHeight,
			Hash:                 transaction.Hash,
			Version:              transaction.Version,
			Type:                 transaction.Type,
			LockTime:             transaction.LockTime,
			Fee:                  transaction.Fee,
			Info:                 transaction.Info,
			SigPubKey:            transaction.SigPubKey,
			Sig:                  transaction.Sig,
			Proof:                transaction.Proof,
			Metadata:             transaction.Metadata,
			PubKeyLastByteSender: transaction.PubKeyLastByteSender,
		}

		if transaction.TxPrivacy != nil {
			customTokenTransaction := &model.TransactionCustomToken{
				Tx: model.Transaction{
					ShardId:              shard.ShardID,
					ShardHash:            shard.BlockHash,
					ShardHeight:          shard.Height,
					Hash:                 transaction.Hash,
					Version:              transaction.Version,
					Type:                 transaction.TxPrivacy.Tx.Type,
					LockTime:             transaction.LockTime,
					Fee:                  transaction.Fee,
					Info:                 transaction.Info,
					SigPubKey:            transaction.TxPrivacy.Tx.SigPubKey,
					Sig:                  transaction.TxPrivacy.Tx.Sig,
					Proof:                transaction.TxPrivacy.Tx.Proof,
					PubKeyLastByteSender: transaction.TxPrivacy.Tx.PubKeyLastByteSender,
					Metadata:             transaction.Metadata,
				},
				PropertyID:     transaction.TxPrivacy.PropertyID,
				PropertyName:   transaction.TxPrivacy.PropertyName,
				PropertySymbol: transaction.TxPrivacy.PropertySymbol,
				Type:           transaction.TxPrivacy.Type,
				Mintable:       transaction.TxPrivacy.Mintable,
				Amount:         transaction.TxPrivacy.Amount,
			}
			newTransaction.TransactionCustomToken = customTokenTransaction
		}
		transactions = append(transactions, newTransaction)
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
				Value:           input.Value,
				Info:            input.Info,
				TokenID:         input.TokenID,
			}
			if input.PublicKey != nil {
				inputCoin.PublicKey = input.PublicKey.ToBytesS()
			}
			if input.CoinCommitment != nil {
				inputCoin.CoinCommitment = input.CoinCommitment.ToBytesS()
			}
			if input.SNDerivator != nil {
				inputCoin.SNDerivator = input.SNDerivator.ToBytesS()
			}
			if input.SerialNumber != nil {
				inputCoin.SerialNumber = input.SerialNumber.ToBytesS()
			}
			if input.Randomness != nil {
				inputCoin.Randomness = input.Randomness.ToBytesS()
			}
			inputCoins = append(inputCoins, inputCoin)
		}

	}
	return inputCoins
}
func getCrossShardOutputCoinFromShardState(shard *data.Shard) []model.OutputCoin {
	outputCoins := make([]model.OutputCoin, 0, len(shard.OutputCoins))
	for _, output := range shard.OutputCoins {
		if output.ToShardID == shard.ShardID {
			continue
		}
		outputCoin := model.OutputCoin{
			ShardId:          shard.ShardID,
			ShardHash:        shard.BlockHash,
			ShardHeight:      shard.BeaconHeight,
			TransactionHash:  output.TransactionHash,
			Value:            output.Value,
			Info:             output.Info,
			TokenID:          output.TokenID,
			FromShardID:      output.FromShardID,
			ToShardID:        output.ToShardID,
			FromCrossShard:   output.FromCrossShard,
			CrossBlockHash:   output.CrossBlockHash,
			CrossBlockHeight: output.CrossBlockHeight,
			PropertyName:     output.PropertyName,
			PropertySymbol:   output.PropertySymbol,
			Type:             output.Type,
			Mintable:         output.Mintable,
			Amount:           output.Amount,
		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = output.PublicKey.ToBytesS()
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = output.CoinCommitment.ToBytesS()
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = output.SNDerivator.ToBytesS()
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = output.SerialNumber.ToBytesS()
		}
		if output.Randomness != nil {
			outputCoin.Randomness = output.Randomness.ToBytesS()
		}
		outputCoins = append(outputCoins, outputCoin)
	}
	return outputCoins
}
func getOutputCoinForThisShardFromShardState(shard *data.Shard) []model.OutputCoin {
	outputCoins := make([]model.OutputCoin, 0, len(shard.OutputCoins))
	for _, output := range shard.OutputCoins {
		if output.ToShardID != shard.ShardID {
			continue
		}
		outputCoin := model.OutputCoin{
			ShardId:          shard.ShardID,
			ShardHash:        shard.BlockHash,
			ShardHeight:      shard.BeaconHeight,
			TransactionHash:  output.TransactionHash,
			Value:            output.Value,
			Info:             output.Info,
			TokenID:          output.TokenID,
			FromShardID:      output.FromShardID,
			ToShardID:        output.ToShardID,
			FromCrossShard:   output.FromCrossShard,
			CrossBlockHash:   output.CrossBlockHash,
			CrossBlockHeight: output.CrossBlockHeight,
			PropertyName:     output.PropertyName,
			PropertySymbol:   output.PropertySymbol,
			Type:             output.Type,
			Mintable:         output.Mintable,
			Amount:           output.Amount,
		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = output.PublicKey.ToBytesS()
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = output.CoinCommitment.ToBytesS()
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = output.SNDerivator.ToBytesS()
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = output.SerialNumber.ToBytesS()
		}
		if output.Randomness != nil {
			outputCoin.Randomness = output.Randomness.ToBytesS()
		}
		outputCoins = append(outputCoins, outputCoin)
	}
	return outputCoins
}

func getCommitmentFromShardState(shard *data.Shard) []model.Commitment {
	commitments := make([]model.Commitment, 0, len(shard.Commitments))

	for _, commitment := range shard.Commitments {
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

func GetTokenStateFromShardState(shard *data.Shard) model.TokenState {
	tokenState := model.TokenState{
		ShardID:     shard.ShardID,
		ShardHash:   shard.BlockHash,
		ShardHeight: shard.Height,
	}
	tokenInformations := make([]model.TokenInformation, 0, len(shard.TokenState))

	for _, token := range shard.TokenState {
		tokenInformations = append(tokenInformations, model.TokenInformation{
			TokenID:        token.TokenID,
			PropertyName:   token.PropertyName,
			PropertySymbol: token.PropertySymbol,
			TokenType:      token.TokenType,
			Mintable:       token.Mintable,
			Amount:         token.Amount,
			Info:           token.Info,
			InitTx:         token.InitTx,
			Txs:            token.Txs,
		})
	}
	tokenState.TokenInformations = tokenInformations
	return tokenState
}
