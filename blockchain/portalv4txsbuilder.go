package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalAcceptedShieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalAcceptedShieldingRequestTx] Starting...")
	contentBytes := []byte(contentStr)
	var acceptedShieldingReq metadata.PortalShieldingRequestContent
	err := json.Unmarshal(contentBytes, &acceptedShieldingReq)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal custodian deposit content: %+v", err)
		return nil, nil
	}
	if acceptedShieldingReq.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, acceptedShieldingReq.ShardID)
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
		Logger.log.Errorf("ERROR: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := acceptedShieldingReq.MintingAmount
	tokenID, _ := new(common.Hash).NewHashFromStr(acceptedShieldingReq.TokenID)

	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         receiveAmt,
		PaymentAddress: receiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   issuingAcceptedInst.IncTokenName,
		// PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:      receiveAmt,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	txStateDB := curView.GetCopiedTransactionStateDB()
	featureStateDB := beaconState.GetBeaconFeatureStateDB()
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(
			producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			txStateDB,
			meta,
			false,
			false,
			shardID,
			nil,
			featureStateDB,
		),
	)
	if initErr != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing shielding response tx: %+v", initErr)
		return nil, nil
	}
	return resTx, nil
}

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalRejectedUnshieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRejectedUnshieldingRequestTx] Starting...")
	contentBytes := []byte(contentStr)
	var rejectedUnshieldingReq metadata.PortalUnshieldRequestContent
	err := json.Unmarshal(contentBytes, &rejectedUnshieldingReq)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal v4 unshield request content: %+v", err)
		return nil, nil
	}
	if rejectedUnshieldingReq.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, rejectedUnshieldingReq.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalV4UnshieldResponse(
		"rejected",
		rejectedUnshieldingReq.TxReqID,
		rejectedUnshieldingReq.IncAddressStr,
		rejectedUnshieldingReq.UnshieldAmount,
		rejectedUnshieldingReq.TokenID,
		metadata.PortalV4UnshieldingRequestMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(rejectedUnshieldingReq.IncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := rejectedUnshieldingReq.UnshieldAmount
	tokenID, _ := new(common.Hash).NewHashFromStr(rejectedUnshieldingReq.TokenID)

	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         receiveAmt,
		PaymentAddress: receiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], tokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   issuingAcceptedInst.IncTokenName,
		// PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:      receiveAmt,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	txStateDB := curView.GetCopiedTransactionStateDB()
	featureStateDB := beaconState.GetBeaconFeatureStateDB()
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(
			producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			txStateDB,
			meta,
			false,
			false,
			shardID,
			nil,
			featureStateDB,
		),
	)
	if initErr != nil {
		Logger.log.Errorf("ERROR: an error occured while initializing shielding response tx: %+v", initErr)
		return nil, nil
	}
	return resTx, nil
}
