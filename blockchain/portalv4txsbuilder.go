package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalAcceptedShieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[buildPortalAcceptedShieldingRequestTx] Starting...")

	contentBytes := []byte(contentStr)
	var acceptedShieldingReq metadata.PortalShieldingRequestContent
	err := json.Unmarshal(contentBytes, &acceptedShieldingReq)
	if err != nil {
		Logger.log.Errorf("[buildPortalAcceptedShieldingRequestTx]: an error occured while unmarshaling portal custodian deposit content: %+v", err)
		return nil, nil
	}
	if acceptedShieldingReq.ShardID != shardID {
		Logger.log.Errorf("[buildPortalAcceptedShieldingRequestTx]: ShardID unexpected expect %v, but got %+v", shardID, acceptedShieldingReq.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalShieldingResponse(
		"accepted",
		acceptedShieldingReq.TxReqID,
		acceptedShieldingReq.IncogAddressStr,
		acceptedShieldingReq.MintingAmount,
		acceptedShieldingReq.TokenID,
		metadata.PortalV4ShieldingResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(acceptedShieldingReq.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("[buildPortalAcceptedShieldingRequestTx]: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         acceptedShieldingReq.MintingAmount,
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
	}

	tokenIDPointer, _ := new(common.Hash).NewHashFromStr(acceptedShieldingReq.TokenID)
	tokenID := *tokenIDPointer
	if tokenID == common.PRVCoinID {
		Logger.log.Errorf("[buildPortalAcceptedShieldingRequestTx]: cannot minting PRV in shield request")
		return nil, errors.New("[buildPortalAcceptedShieldingRequestTx]: cannot mint PRV in shield request")
	}

	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
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
		metadata.PortalV4UnshieldingResponseMeta,
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
