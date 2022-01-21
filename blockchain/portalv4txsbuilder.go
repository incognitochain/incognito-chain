package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction"
)

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalAcceptedShieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	prefix := fmt.Sprintf("[buildPortalAcceptedShieldingRequestTx]")
	Logger.log.Infof("%v ContentStr: %v\n", contentStr)
	contentBytes := []byte(contentStr)
	var acceptedShieldingReq metadata.PortalShieldingRequestContent
	err := json.Unmarshal(contentBytes, &acceptedShieldingReq)
	if err != nil {
		Logger.log.Errorf("%v an error occurred while unmarshalling instruction's content: %v\n",
			prefix, err)
		return nil, fmt.Errorf("%v cannot unmarshal instruction's content: %v", prefix, err)
	}
	if acceptedShieldingReq.ShardID != shardID {
		Logger.log.Errorf("%v expected shardID %v, got %v\n", prefix, shardID, acceptedShieldingReq.ShardID)
		return nil, fmt.Errorf("%v expected shardID %v, got %v", prefix, shardID, acceptedShieldingReq.ShieldingUTXO)
	}

	meta := metadata.NewPortalShieldingResponse(
		"accepted",
		acceptedShieldingReq.TxReqID,
		acceptedShieldingReq.Receiver,
		acceptedShieldingReq.MintingAmount,
		acceptedShieldingReq.TokenID,
		metadataCommon.PortalV4ShieldingResponseMeta,
	)

	tokenIDPointer, _ := new(common.Hash).NewHashFromStr(acceptedShieldingReq.TokenID)
	tokenID := *tokenIDPointer
	if tokenID == common.PRVCoinID {
		Logger.log.Errorf("%v cannot minting PRV in shield request\n", prefix)
		return nil, fmt.Errorf("%v cannot mint PRV in shield request", prefix)
	}

	txParam := transaction.TxSalaryOutputParams{
		Amount:  acceptedShieldingReq.MintingAmount,
		TokenID: &tokenID,
	}
	keyWallet, err := wallet.Base58CheckDeserialize(acceptedShieldingReq.Receiver)
	if err == nil { // receiver is a payment address
		txParam.ReceiverAddress = &keyWallet.KeySet.PaymentAddress
	} else { // receiver is an OTAReceiver
		otaReceiver := new(privacy.OTAReceiver)
		err = otaReceiver.FromString(acceptedShieldingReq.Receiver)
		if err != nil {
			return nil, fmt.Errorf("parseOTA receiver from %v error: %v", acceptedShieldingReq.Receiver, err)
		}
		txParam.TxRandom = &otaReceiver.TxRandom
		txParam.PublicKey = &otaReceiver.PublicKey
	}

	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalRefundedUnshieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Infof("[buildPortalRefundedUnshieldingRequestTx] Starting...")

	// parse instruction content
	contentBytes := []byte(contentStr)
	var unshieldRequest metadata.PortalUnshieldRequestContent
	err := json.Unmarshal(contentBytes, &unshieldRequest)
	if err != nil {
		Logger.log.Errorf("[buildPortalRefundedUnshieldingRequestTx]: an error occured while unmarshaling portal v4 unshield request content: %+v", err)
		return nil, nil
	}

	// check shardID
	if unshieldRequest.ShardID != shardID {
		Logger.log.Errorf("[buildPortalRefundedUnshieldingRequestTx]: ShardID unexpected expect %v, but got %+v", shardID, unshieldRequest.ShardID)
		return nil, nil
	}

	// create metadata UnshieldResponse
	meta := metadata.NewPortalV4UnshieldResponse(
		"refunded",
		unshieldRequest.TxReqID,
		unshieldRequest.OTAPubKeyStr,
		unshieldRequest.TxRandomStr,
		unshieldRequest.UnshieldAmount,
		unshieldRequest.TokenID,
		metadataCommon.PortalV4UnshieldingResponseMeta,
	)

	// init salary tx
	tokenID, err := common.Hash{}.NewHashFromStr(unshieldRequest.TokenID)
	if err != nil {
		Logger.log.Errorf("[buildPortalRefundedUnshieldingRequestTx]: an error occured while converting tokenid to hash: %+v", err)
		return nil, err
	}
	publicKey, txRandom, err := coin.ParseOTAInfoFromString(unshieldRequest.OTAPubKeyStr, unshieldRequest.TxRandomStr)
	if err != nil {
		Logger.log.Errorf("[buildPortalRefundedUnshieldingRequestTx]: an error occured while parse ota address: %+v", err)
		return nil, err
	}
	var txParam transaction.TxSalaryOutputParams
	txParam = transaction.TxSalaryOutputParams{Amount: unshieldRequest.UnshieldAmount, ReceiverAddress: nil, PublicKey: publicKey, TxRandom: txRandom, TokenID: tokenID, Info: []byte{}}

	resTx, err := txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), func(c privacy.Coin) metadata.Metadata {
		return meta
	})
	if err != nil {
		Logger.log.Errorf("[buildPortalRefundedUnshieldingRequestTx]: an error occured while initializing refund unshielding response tx: %+v", err)
		return nil, nil
	}
	Logger.log.Info("[buildPortalRefundedUnshieldingRequestTx] Create response tx successfully.")
	return resTx, nil
}
