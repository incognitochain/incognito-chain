package transaction

import (
	"fmt"
	"errors"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
)

func TestResignTxV1(txv1_generic metadata.Transaction){
	txv1, ok := txv1_generic.(*TxVersion1)
	if !ok{
		panic("Error when casting")
	}
	txv1.cachedHash = nil
	txv1.SetSig(nil)
	txv1.SetSigPubKey(nil)
	err := txv1.sign()
	if err!=nil{
		// if it fails, something's wrong
		fmt.Printf("Error when resigning : %v\n",err)
	}
}

func (tx *TxVersion2) InitTestOldOTA(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyInitParams)
	if !ok {
		return errors.New("params of tx Init is not TxPrivacyInitParam")
	}

	Logger.Log.Debugf("CREATING TX........\n")
	if err := validateTxParams(params); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.initializeTxAndParams(params); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.isNonPrivacyNonInput(params); check {
		return err
	}

	outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins
	clonedPK := &operation.Point{}
	clonedPK.FromBytes(inputCoins[0].GetPublicKey().ToBytes())
	outputCoins[0].SetPublicKey(clonedPK)

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
		if err := tx.signMetadata(params.senderSK); err != nil {
			Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}
	err = tx.signOnMessage(inputCoins, outputCoins, params, tx.Hash()[:])
	return err
	// return nil
}