package storage

import (
	"context"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"time"
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
	pdeShares := getPDEShareFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetPDEShareStorer().StorePDEShare(ctx, pdeShares); err != nil {
		panic(err)
		return err
	}

	pdePoolPairs := getPDEPoolForPairStateFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetPDEPoolForPairStateStorer().StorePDEPoolForPairState(ctx, pdePoolPairs); err != nil {
		panic(err)
		return err
	}


	pdeTradingFees := getPDETradingFeeFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetPDETradingFeeStorer().StorePDETradingFee(ctx, pdeTradingFees); err != nil {
		panic(err)
		return err
	}

	waitingPDEContributionStates := getWaitingPDEContributionStateFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetWaitingPDEContributionStorer().StoreWaitingPDEContribution(ctx, waitingPDEContributionStates); err != nil {
		panic(err)
		return err
	}


	//Portal
	custodians := getCustodianFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetCustodianStorer().StoreCustodian(ctx, custodians); err != nil {
		panic(err)
		return err
	}

	waitingPortingRequests := getWaitingPortingRequestFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetWaitingPortingRequestStorer().StoreWaitingPortingRequest(ctx, waitingPortingRequests); err != nil {
		panic(err)
		return err
	}

	matchedRedeemRequests := getMatchedRedeemRequestFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetMatchedRedeemRequestStorer().StoreMatchedRedeemRequest(ctx, matchedRedeemRequests); err != nil {
		panic(err)
		return err
	}

	waitingRedeemRequests := getWaitingRedeemRequestFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetWaitingRedeemRequestStorer().StoreWaitingRedeemRequest(ctx, waitingRedeemRequests); err != nil {
		panic(err)
		return err
	}

	finalExchangeRates := getFinalExchangeRatesFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetFinalExchangeRatesStorer().StoreFinalExchangeRates(ctx, finalExchangeRates); err != nil {
		panic(err)
		return err
	}

	lockedCollaterals := getLockedCollateralFromBeaconState(beacon)
	if err := GetDBDriver(MONGODB).GetLockedCollateralStorer().StoreLockedCollateral(ctx, lockedCollaterals); err != nil {
		panic(err)
		return err
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

func getPDEShareFromBeaconState(beacon *data.Beacon) model.PDEShare {
	pdeShareInfos := make([]model.PDEShareInfo, 0, len(beacon.PDEShare))
	for _, share := range beacon.PDEShare {
		pdeShareInfos = append(pdeShareInfos, model.PDEShareInfo{
			Token1ID:           share.Token1ID,
			Token2ID:           share.Token2ID,
			ContributorAddress: share.ContributorAddress,
			Amount:             share.Amount,
		})
	}
	return model.PDEShare{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		PDEShareInfo:       pdeShareInfos,
	}
}
func getWaitingPDEContributionStateFromBeaconState(beacon *data.Beacon) model.WaitingPDEContribution {
	waitingPDEContributionInfos := make([]model.WaitingPDEContributionInfo, 0, len(beacon.WaitingPDEContributionState))
	for _, waiting := range beacon.WaitingPDEContributionState {
		waitingPDEContributionInfos = append(waitingPDEContributionInfos, model.WaitingPDEContributionInfo{
			PairID:             waiting.PairID,
			ContributorAddress: waiting.ContributorAddress,
			TokenID:            waiting.TokenID,
			Amount:             waiting.Amount,
			TXReqID:            waiting.TXReqID,
		})
	}
	return model.WaitingPDEContribution{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		WaitingPDEContributionInfo: waitingPDEContributionInfos,
	} 
}

func getPDETradingFeeFromBeaconState(beacon *data.Beacon) model.PDETradingFee {
	pdeTradingFeeInfos := make([]model.PDETradingFeeInfo, 0, len(beacon.PDETradingFee))
	for _, pdeTradingFee := range beacon.PDETradingFee {
		pdeTradingFeeInfos = append(pdeTradingFeeInfos, model.PDETradingFeeInfo{
			Token1ID:           pdeTradingFee.Token1ID,
			Token2ID:           pdeTradingFee.Token2ID,
			ContributorAddress: pdeTradingFee.ContributorAddress,
			Amount:             pdeTradingFee.Amount,
		})
	}
	return model.PDETradingFee{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		PDETradingFeeInfo:  pdeTradingFeeInfos,
	}
}

func getPDEPoolForPairStateFromBeaconState(beacon *data.Beacon) model.PDEPoolForPair {
	pdeFoolForPairInfos := make([]model.PDEPoolForPairInfo, 0, len(beacon.PDEPoolPair))
	for _, pdeFoolForPair := range beacon.PDEPoolPair {
		pdeFoolForPairInfos = append(pdeFoolForPairInfos, model.PDEPoolForPairInfo{
			Token1ID:        pdeFoolForPair.Token1ID,
			Token1PoolValue: pdeFoolForPair.Token1PoolValue,
			Token2ID:        pdeFoolForPair.Token2ID,
			Token2PoolValue: pdeFoolForPair.Token2PoolValue,
		})
	}
	return model.PDEPoolForPair{
		BeaconBlockHash: beacon.BlockHash,
		BeaconEpoch:     beacon.Epoch,
		BeaconHeight:    beacon.Height,
		BeaconTime:      beacon.Time,
		PDEPoolForPairInfo: nil,
	}
}

func getCustodianFromBeaconState(beacon *data.Beacon) model.Custodian {
	custodianInfos := make([]model.CustodianInfo, 0, len(beacon.Custodian))
	for _, custodian := range beacon.Custodian {
		custodianInfos = append(custodianInfos, model.CustodianInfo{
			IncognitoAddress:       custodian.IncognitoAddress,
			TotalCollateral:        custodian.TotalCollateral,
			FreeCollateral:         custodian.FreeCollateral,
			HoldingPubTokens:       custodian.HoldingPubTokens,
			LockedAmountCollateral: custodian.LockedAmountCollateral,
			RemoteAddresses:        custodian.RemoteAddresses,
			RewardAmount:           custodian.RewardAmount,
		})
	}
	return model.Custodian{
		BeaconBlockHash:        beacon.BlockHash,
		BeaconEpoch:            beacon.Epoch,
		BeaconHeight:           beacon.Height,
		BeaconTime:             beacon.Time,
		CustodianInfo:          custodianInfos,
	}
}

func getWaitingPortingRequestFromBeaconState(beacon *data.Beacon) model.WaitingPortingRequest {
	waitingPortingRequestInfos := make([]model.WaitingPortingRequestInfo, 0, len(beacon.WaitingPortingRequest))
	for _, w := range beacon.WaitingPortingRequest {
		waitingPortingRequestInfos = append(waitingPortingRequestInfos, model.WaitingPortingRequestInfo{
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
	return model.WaitingPortingRequest{
		BeaconBlockHash:     beacon.BlockHash,
		BeaconEpoch:         beacon.Epoch,
		BeaconHeight:        beacon.Height,
		BeaconTime:          beacon.Time,
		WaitingPortingRequestInfo: waitingPortingRequestInfos,
	}
}

func getFinalExchangeRatesFromBeaconState(beacon *data.Beacon) model.FinalExchangeRate {
	finalExchangeRateInfos := make([]model.FinalExchangeRateInfo, 0, len(beacon.FinalExchangeRates.Rates))
	for key, amount := range beacon.FinalExchangeRates.Rates {
		finalExchangeRateInfos = append(finalExchangeRateInfos, model.FinalExchangeRateInfo{
			Amount:          amount.Amount,
			TokenID:         key,
		})
	}
	return model.FinalExchangeRate{
		BeaconBlockHash: beacon.BlockHash,
		BeaconEpoch:     beacon.Epoch,
		BeaconHeight:    beacon.Height,
		BeaconTime:      beacon.Time,
		FinalExchangeRateInfo: finalExchangeRateInfos,
	}
}

func getMatchedRedeemRequestFromBeaconState(beacon *data.Beacon) model.RedeemRequest {
	redeemRequestInfos := make([]model.RedeemRequestInfo, 0, len(beacon.MatchedRedeemRequest))
	for _, matchedRedeem := range beacon.MatchedRedeemRequest {
		redeemRequestInfos = append(redeemRequestInfos, model.RedeemRequestInfo{
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
	return model.RedeemRequest{
		BeaconBlockHash:       beacon.BlockHash,
		BeaconEpoch:           beacon.Epoch,
		BeaconHeight:          beacon.Height,
		BeaconTime:            beacon.Time,
		RedeemRequestInfo: redeemRequestInfos,
	}
}

func getWaitingRedeemRequestFromBeaconState(beacon *data.Beacon) model.RedeemRequest {
	redeemRequestInfos := make([]model.RedeemRequestInfo, 0, len(beacon.WaitingRedeemRequest))
	for _, waitingRedeem := range beacon.WaitingRedeemRequest {
		redeemRequestInfos = append(redeemRequestInfos, model.RedeemRequestInfo{
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
	return model.RedeemRequest{
		BeaconBlockHash:       beacon.BlockHash,
		BeaconEpoch:           beacon.Epoch,
		BeaconHeight:          beacon.Height,
		BeaconTime:            beacon.Time,
		RedeemRequestInfo: redeemRequestInfos,
	}
}

func getLockedCollateralFromBeaconState(beacon *data.Beacon) model.LockedCollateral {
	lockedCollateralInfos := make([]model.LockedCollateralInfo, 0, len(beacon.LockedCollateralState.LockedCollateralDetail))
	for key, lockedDetail := range beacon.LockedCollateralState.LockedCollateralDetail {
		lockedCollateralInfos = append(lockedCollateralInfos, model.LockedCollateralInfo{
			TotalLockedCollateralForRewards: beacon.LockedCollateralState.TotalLockedCollateralForRewards,
			CustodianAddress:                key,
			Amount:                          lockedDetail,
		})
	}
	return model.LockedCollateral{
		BeaconBlockHash:                 beacon.BlockHash,
		BeaconEpoch:                     beacon.Epoch,
		BeaconHeight:                    beacon.Height,
		BeaconTime:                      beacon.Time,
		LockedCollateralInfo: lockedCollateralInfos,
	}
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

	tokenState := GetTokenStateFromShardState(shard)
	if err := GetDBDriver(MONGODB).GetTokenStateStorer().StoreTokenState(ctx, tokenState); err != nil {
		panic(err)
		return err
	}

	rewardState := GetRewardStateFromShardState(shard)
	if err := GetDBDriver(MONGODB).GetCommitteeRewardStateStorer().StoreCommitteeRewardState(ctx, rewardState); err != nil {
		panic(err)
		return err
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
		ShardTxRoot:	        shard.ShardTxRoot,
		CrossTransactionRoot:   shard.CrossTransactionRoot,
		InstructionsRoot:       shard.InstructionsRoot,
		CommitteeRoot:          shard.CommitteeRoot,
		PendingValidatorRoot:   shard.PendingValidatorRoot,
		StakingTxRoot:          shard.StakingTxRoot,
		InstructionMerkleRoot:  shard.InstructionMerkleRoot,
		TotalTxsFee:            shard.TotalTxsFee,
		Time:                   shard.Time,
		TxHashes:               shard.TxHashes,
		Txs:                    shard.Txs,
		BlockProducer:          shard.BlockProducer,
		BlockProducerPubKeyStr: shard.BlockProducerPubKeyStr,
		Proposer: 				shard.Proposer,
		ProposeTime: 			shard.ProposeTime,
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
			Image:                 "",
			IsPrivacy:             transaction.IsPrivacy,
			TxSize:				  transaction.TxSize,
			Index:                transaction.Index,
			Hash:                 transaction.Hash,
			Version:              transaction.Version,
			Type:                 transaction.Type,
			LockTime:             time.Unix(transaction.LockTime, 0).Format(common.DateOutputFormat),
			Fee:                  transaction.Fee,
			Info:                 string(transaction.Info),
			SigPubKey:            base58.Base58Check{}.Encode(transaction.SigPubKey, 0x0),
			Sig:                  base58.Base58Check{}.Encode(transaction.Sig, 0x0),
			PubKeyLastByteSender: transaction.PubKeyLastByteSender,
			InputCoinPubKey: transaction.InputCoinPubKey,
			IsInBlock: true,
			IsInMempool: false,
		}
		newTransaction.ProofDetail, newTransaction.Proof = 	getProofDetail(transaction)
		newTransaction.CustomTokenData =  ""
		if transaction.Metadata != nil {
			metaData, _ := json.MarshalIndent(transaction.Metadata, "", "\t")
			newTransaction.Metadata = string(metaData)
		}
		if transaction.TxPrivacy != nil {
			newTransaction.PrivacyCustomTokenID = transaction.TxPrivacy.PropertyID
			newTransaction.PrivacyCustomTokenName = transaction.TxPrivacy.PropertyName
			newTransaction.PrivacyCustomTokenSymbol = transaction.TxPrivacy.PropertySymbol
			newTransaction.PrivacyCustomTokenData = transaction.PrivacyCustomTokenData
			newTransaction.PrivacyCustomTokenIsPrivacy = transaction.TxPrivacy.Tx.IsPrivacy
			newTransaction.PrivacyCustomTokenFee = transaction.TxPrivacy.Tx.Fee
			newTransaction.PrivacyCustomTokenProofDetail, newTransaction.PrivacyCustomTokenProof = getProofDetail(transaction.TxPrivacy.Tx)
		}
		transactions = append(transactions, newTransaction)
	}
	return transactions
}

func getProofDetail (normalTx *data.Transaction) (jsonresult.ProofDetail, *string) {
	if normalTx.Proof == nil {
		return jsonresult.ProofDetail{}, nil
	}
	b, _:= normalTx.Proof.MarshalJSON()
	proof := string(b)
	return jsonresult.ProofDetail{
		InputCoins:  getProofDetailInputCoin(normalTx.Proof),
		OutputCoins: getProofDetailOutputCoin(normalTx.Proof),
	}, &proof
}

func getProofDetailInputCoin(proof *zkp.PaymentProof) []*jsonresult.CoinDetail {
	inputCoins := make([]*jsonresult.CoinDetail, 0)
	for _, input := range proof.GetInputCoins() {
		in := jsonresult.CoinDetail{
			CoinDetails: jsonresult.Coin{},
		}
		if input.CoinDetails != nil {
			in.CoinDetails.Value = input.CoinDetails.GetValue()
			in.CoinDetails.Info = base58.Base58Check{}.Encode(input.CoinDetails.GetInfo(), 0x0)
			if input.CoinDetails.GetCoinCommitment() != nil {
				in.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(input.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetRandomness() != nil {
				in.CoinDetails.Randomness = *input.CoinDetails.GetRandomness()
			}
			if input.CoinDetails.GetSNDerivator() != nil {
				in.CoinDetails.SNDerivator = *input.CoinDetails.GetSNDerivator()
			}
			if input.CoinDetails.GetSerialNumber() != nil {
				in.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(input.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetPublicKey() != nil {
				in.CoinDetails.PublicKey = base58.Base58Check{}.Encode(input.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
		}
		inputCoins = append(inputCoins, &in)
	}
	return inputCoins
}

func getProofDetailOutputCoin(proof *zkp.PaymentProof) []*jsonresult.CoinDetail {
	outputCoins := make([]*jsonresult.CoinDetail, 0)
	for _, output := range proof.GetOutputCoins() {
		out := jsonresult.CoinDetail{
			CoinDetails: jsonresult.Coin{},
		}
		if output.CoinDetails != nil {
			out.CoinDetails.Value = output.CoinDetails.GetValue()
			out.CoinDetails.Info = base58.Base58Check{}.Encode(output.CoinDetails.GetInfo(), 0x0)
			if output.CoinDetails.GetCoinCommitment() != nil {
				out.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(output.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetRandomness() != nil {
				out.CoinDetails.Randomness = *output.CoinDetails.GetRandomness()
			}
			if output.CoinDetails.GetSNDerivator() != nil {
				out.CoinDetails.SNDerivator = *output.CoinDetails.GetSNDerivator()
			}
			if output.CoinDetails.GetSerialNumber() != nil {
				out.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(output.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetPublicKey() != nil {
				out.CoinDetails.PublicKey = base58.Base58Check{}.Encode(output.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
			if output.CoinDetailsEncrypted != nil {
				out.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), 0x0)
			}
		}
		outputCoins = append(outputCoins , &out)
	}
	return outputCoins
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
				Info:            string(input.Info),
				TokenID:         input.TokenID,
			}
			if input.PublicKey != nil {
				inputCoin.PublicKey =   base58.Base58Check{}.Encode(input.PublicKey.ToBytesS(), common.ZeroByte)
			}
			if input.CoinCommitment != nil {
				inputCoin.CoinCommitment = base58.Base58Check{}.Encode(input.CoinCommitment.ToBytesS(), common.ZeroByte)
			}
			if input.SNDerivator != nil {
				inputCoin.SNDerivator = base58.Base58Check{}.Encode(input.SNDerivator.ToBytesS(), common.ZeroByte)
			}
			if input.SerialNumber != nil {
				inputCoin.SerialNumber = base58.Base58Check{}.Encode(input.SerialNumber.ToBytesS(), common.ZeroByte)
			}
			if input.Randomness != nil {
				inputCoin.Randomness = base58.Base58Check{}.Encode(input.Randomness.ToBytesS(), common.ZeroByte)
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
			Info:             string(output.Info),
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
			LockTime:		  output.LockTime,
			TransactionMemo: string(output.TransactionMemo),

		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = base58.Base58Check{}.Encode(output.PublicKey.ToBytesS(), common.ZeroByte)
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = base58.Base58Check{}.Encode(output.CoinCommitment.ToBytesS(), common.ZeroByte)
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = base58.Base58Check{}.Encode(output.SNDerivator.ToBytesS(), common.ZeroByte)
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = base58.Base58Check{}.Encode(output.SerialNumber.ToBytesS(), common.ZeroByte)
		}
		if output.Randomness != nil {
			outputCoin.Randomness = base58.Base58Check{}.Encode(output.Randomness.ToBytesS(), common.ZeroByte)
		}
		if output.CoinDetailsEncrypted != nil {
			outputCoin.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
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
			Info:             string(output.Info),
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
			LockTime:		  output.LockTime,
			TransactionMemo: string(output.TransactionMemo),

		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = base58.Base58Check{}.Encode(output.PublicKey.ToBytesS(), common.ZeroByte)
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = base58.Base58Check{}.Encode(output.CoinCommitment.ToBytesS(), common.ZeroByte)
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = base58.Base58Check{}.Encode(output.SNDerivator.ToBytesS(), common.ZeroByte)
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = base58.Base58Check{}.Encode(output.SerialNumber.ToBytesS(), common.ZeroByte)
		}
		if output.Randomness != nil {
			outputCoin.Randomness = base58.Base58Check{}.Encode(output.Randomness.ToBytesS(), common.ZeroByte)
		}
		if output.CoinDetailsEncrypted != nil {
			outputCoin.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
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
			Commitment:      base58.Base58Check{}.Encode(commitment.Commitment,common.ZeroByte),
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
	tokenState.Token = tokenInformations
	return tokenState
}

func GetRewardStateFromShardState(shard *data.Shard) model.CommitteeRewardState {
	rewardsState := model.CommitteeRewardState{
		ShardID:     shard.ShardID,
		ShardHash:   shard.BlockHash,
		ShardHeight: shard.Height,
	}
	rewards := make([]model.CommitteeReward, 0, 2000)

	for address, token := range shard.CommitteeRewardState {
		for token, amount := range token {
			rewards = append(rewards, model.CommitteeReward{
				Address: address,
				TokenId: token,
				Amount:  amount,
			})
		}

	}
	rewardsState.CommitteeReward = rewards
	return rewardsState
}
