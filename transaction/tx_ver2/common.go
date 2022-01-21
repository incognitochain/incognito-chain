package tx_ver2

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

var (
	tokenPrivAttributeMap = initTokenPrivacyAttributes()
)

func initTokenPrivacyAttributes() map[common.Hash]privacy.TokenAttributes {
	result := make(map[common.Hash]privacy.TokenAttributes)
	result[common.PdexAccessCoinID] = privacy.TokenAttributes{Private: false, BurnOnly: true}
	return result
}

func getDerivableInputFromSigPubKey(rawPubkey []byte, tokenID common.Hash, txHash *common.Hash, shardID byte, db *statedb.StateDB) (*privacy.Point, error) {
	ringPubkey := new(SigPubKey)
	if err := ringPubkey.SetBytes(rawPubkey); err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("invalid SigPubKey in tx %s, token %s", txHash.String(), tokenID.String()))
	}
	ringSize := len(ringPubkey.Indexes)
	if ringSize != 1 || len(ringPubkey.Indexes[0]) != 1 {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("cannot identify burn input in tx %s, token %s with ring size %d, input length %d", txHash.String(), tokenID.String(), ringSize, len(ringPubkey.Indexes[0])))
	}
	rawCoin, err := statedb.GetOTACoinByIndex(db, tokenID, ringPubkey.Indexes[0][0].Uint64(), shardID)
	if err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, err)
	}
	c := new(privacy.CoinV2)
	if err := c.SetBytes(rawCoin); err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, err)
	}
	p := c.GetPublicKey()
	if p == nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("coin from db must have public key"))
	}
	return p, nil
}

// validateNonPrivateTransfer verifies TXs with identifiable inputs, based on that input's token attributes
func validateNonPrivateTransfer(tokenID common.Hash, proof privacy.Proof) (bool, error) {
	att, found := tokenPrivAttributeMap[tokenID]
	if found && att.BurnOnly {
		// a burn-only token can only be burned
		for _, outcoin := range proof.GetOutputCoins() {
			receiverPk := outcoin.GetPublicKey().ToBytesS()
			if !common.IsPublicKeyBurningAddress(receiverPk) {
				return false, fmt.Errorf("cannot transfer burn-only token %v", tokenID)
			}
		}
	}
	return true, nil
}

// ringContainsNonPrivacyToken detects a non-private coin in a ring signature
func ringContainsNonPrivacyToken(ringCoins [][]*privacy.CoinV2) (bool, bool, *common.Hash, *privacy.CoinV2) {
	assetTagMap := privacy.MapPlainAssetTags(tokenPrivAttributeMap)
	for _, row := range ringCoins {
		for _, c := range row {
			tokenID, _ := c.GetTokenId(nil, assetTagMap) // error ignored; only relevant when keySet is present
			if tokenID != nil {
				att := tokenPrivAttributeMap[*tokenID]
				if !att.Private {
					return true, att.BurnOnly, tokenID, c
				}
			}
		}
	}
	return false, false, nil, nil
}
