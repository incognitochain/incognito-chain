package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (blockGenerator *BlockGenerator) buildPortalAcceptedShieldingRequestTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	featureStateDB *statedb.StateDB,
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
	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         acceptedShieldingReq.MintingAmount,
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
	}

	tokenIDPointer, _ := new(common.Hash).NewHashFromStr(acceptedShieldingReq.TokenID)
	tokenID := *tokenIDPointer
	if tokenID == common.PRVCoinID {
		Logger.log.Errorf("cannot minting PRV in shield request")
		return nil, errors.New("cannot mint PRV in shield request")
	}

	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, shardView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalAcceptedShieldingRequestTx builds response tx for the shielding request tx with status "accepted"
// mints pToken to return to user
func (curView *ShardBestState) buildPortalRefundedUnshieldingRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRefundedUnshieldingRequestTx] Starting...")
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
		"refunded",
		rejectedUnshieldingReq.TxReqID,
		rejectedUnshieldingReq.IncAddressStr,
		rejectedUnshieldingReq.UnshieldAmount,
		rejectedUnshieldingReq.TokenID,
		metadata.PortalV4UnshieldingResponseMeta,
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
