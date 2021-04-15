package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"
	"strconv"
)

type WithDrawRewardResponse struct {
	MetadataBase
	TxRequest       *common.Hash
	TokenID         common.Hash
	RewardPublicKey []byte
	SharedRandom    []byte
	Version         int
}

func NewWithDrawRewardResponse(metaRequest *WithDrawRewardRequest, reqID *common.Hash) (*WithDrawRewardResponse, error) {
	metadataBase := MetadataBase{
		Type: WithDrawRewardResponseMeta,
	}
	result := &WithDrawRewardResponse{
		MetadataBase:    metadataBase,
		TxRequest:       reqID,
		TokenID:         metaRequest.TokenID,
		RewardPublicKey: metaRequest.PaymentAddress.Pk[:],
	}
	result.Version = metaRequest.Version
	if ok, err := common.SliceExists(AcceptedWithdrawRewardRequestVersion, result.Version); !ok || err != nil {
		return nil, errors.Errorf("Invalid version %d", result.Version)
	}
	return result, nil
}

func (withDrawRewardResponse WithDrawRewardResponse) Hash() *common.Hash {
	if withDrawRewardResponse.Version == 1 {
		if withDrawRewardResponse.TxRequest == nil {
			return &common.Hash{}
		}
		bArr := append(withDrawRewardResponse.TxRequest.GetBytes(), withDrawRewardResponse.TokenID.GetBytes()...)
		version := strconv.Itoa(withDrawRewardResponse.Version)
		if len(withDrawRewardResponse.SharedRandom) != 0 {
			bArr = append(bArr, withDrawRewardResponse.SharedRandom...)
		}
		if len(withDrawRewardResponse.RewardPublicKey) != 0 {
			bArr = append(bArr, withDrawRewardResponse.RewardPublicKey...)
		}

		bArr = append(bArr, []byte(version)...)
		txResHash := common.HashH(bArr)
		return &txResHash
	} else {
		return withDrawRewardResponse.TxRequest
	}
}

func (withDrawRewardResponse *WithDrawRewardResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	return true
}

func (withDrawRewardResponse *WithDrawRewardResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (withDrawRewardResponse WithDrawRewardResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (withDrawRewardResponse WithDrawRewardResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (withDrawRewardResponse *WithDrawRewardResponse) SetSharedRandom(r []byte) {
	withDrawRewardResponse.SharedRandom = r
}

//func (withDrawRewardResponse WithDrawRewardResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte,
//	tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever,
//	beaconViewRetriever BeaconViewRetriever) (bool, error) {
//
//	if tx.IsPrivacy() {
//		return false, errors.New("This transaction is not private")
//	}
//
//	isMinted, mintCoin, coinID, err := tx.GetTxMintData()
//	//check tx mint
//	if err != nil || !isMinted {
//		return false, errors.Errorf("It is not tx mint with error: %v", err)
//	}
//	//check tokenID
//	if cmp, err := withDrawRewardResponse.TokenID.Cmp(coinID); err != nil || cmp != 0 {
//		return false, errors.Errorf("Token dont match: %v and %v", withDrawRewardResponse.TokenID.String(), coinID.String())
//	}
//
//	//check correct receiver
//	_, _, _, txReq, err := chainRetriever.GetTransactionByHash(*withDrawRewardResponse.TxRequest)
//	if err != nil {
//		return false, errors.Errorf("Cannot get tx request from tx hash %v", withDrawRewardResponse.TxRequest.String())
//	}
//	//check value
//	paymentAddressReq := txReq.GetMetadata().(*WithDrawRewardRequest).PaymentAddress
//	if !bytes.Equal(withDrawRewardResponse.RewardPublicKey, paymentAddressReq.Pk[:]) {
//		return false, errors.Errorf("Wrong reward receiver")
//	}
//
//	pubkeyReqStr := base58.Base58Check{}.Encode(paymentAddressReq.Pk, common.Base58Version)
//	if _, ok := mintData.WithdrawReward[pubkeyReqStr]; ok {
//		return false, errors.New("Verify Miner Mint Tx: Double reward response tx in a block")
//	} else {
//		mintData.WithdrawReward[pubkeyReqStr] = true
//	}
//	rewardAmount, err := statedb.GetCommitteeReward(shardViewRetriever.GetShardRewardStateDB(), pubkeyReqStr, *coinID)
//	fmt.Print("Check Mint Reward Response Valid", mintCoin)
//	fmt.Print("Check Mint Reward Response Valid", paymentAddressReq)
//	fmt.Print("Check Mint Reward Response Valid", withDrawRewardResponse)
//	fmt.Print("Check Mint Reward Response Valid", rewardAmount)
//	if ok := mintCoin.CheckCoinValid(paymentAddressReq, withDrawRewardResponse.SharedRandom, rewardAmount); !ok {
//		return false, errors.New("Mint Coin is invalid")
//	}
//	fmt.Print("Check Mint Reward Response Valid OK OK OK")
//	return true, nil
//}
