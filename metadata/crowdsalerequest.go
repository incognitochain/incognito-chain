package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

// CrowdsaleRequest represents a buying request created by user to send to DCB
type CrowdsaleRequest struct {
	PaymentAddress privacy.PaymentAddress
	SaleID         []byte

	PriceLimit uint64 // max price set by user

	// PriceLimit and Amount is in selling asset: i.e., tx is valid only when price(SellingAsset)/price(BuyingAsset) <= PriceLimit
	LimitSellingAssetPrice bool

	ValidUntil uint64 // in original shard, not beacon or payment shard
	MetadataBase
}

func NewCrowdsaleRequest(csReqData map[string]interface{}) (Metadata, error) {
	errSaver := &ErrorSaver{}
	saleIDStr, okID := csReqData["SaleID"].(string)
	saleID, errSale := hex.DecodeString(saleIDStr)
	priceLimit, okPrice := csReqData["PriceLimit"].(float64)
	validUntil, okValid := csReqData["ValidUntil"].(float64)
	paymentAddressStr, okAddr := csReqData["PaymentAddress"].(string)
	limitSellingAsset, okLimit := csReqData["LimitSellingAssetPrice"].(bool)
	keyWallet, errPayment := wallet.Base58CheckDeserialize(paymentAddressStr)

	if !okID || !okPrice || !okValid || !okAddr || !okLimit {
		return nil, errors.Errorf("Error parsing crowdsale request data")
	}
	if errSaver.Save(errSale, errPayment) != nil {
		return nil, errSaver.Get()
	}

	result := &CrowdsaleRequest{
		PaymentAddress:         keyWallet.KeySet.PaymentAddress,
		SaleID:                 saleID,
		PriceLimit:             uint64(priceLimit),
		ValidUntil:             uint64(validUntil),
		LimitSellingAssetPrice: limitSellingAsset,
	}
	result.Type = CrowdsaleRequestMeta
	return result, nil
}

func (csReq *CrowdsaleRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists and ongoing
	saleData, err := bcr.GetCrowdsaleData(csReq.SaleID)
	fmt.Printf("[db] saleData: %+v\n", saleData)
	if err != nil {
		return false, err
	}

	// TODO(@0xbunyip): get height of beacon chain on new consensus
	beaconHeight := bcr.GetBeaconHeight()
	if beaconHeight >= saleData.EndBlock {
		return false, errors.Errorf("Crowdsale ended")
	}

	// Check if request is still valid
	shardHeight := bcr.GetChainHeight(shardID)
	if shardHeight >= csReq.ValidUntil {
		return false, errors.Errorf("Crowdsale request is not valid anymore")
	}

	// Check if asset is sent to correct address
	if saleData.BuyingAsset.IsEqual(&common.ConstantID) {
		keyWalletBurnAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
		unique, pubkey, _ := txr.GetUniqueReceiver()
		if !unique || !bytes.Equal(pubkey, keyWalletBurnAccount.KeySet.PaymentAddress.Pk[:]) {
			return false, errors.Errorf("Crowdsale request must send CST to Burning address")
		}
	} else {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		unique, pubkey, _ := txr.GetTokenUniqueReceiver()
		fmt.Printf("[db] keywallet and pubkey: \n%x\n%x\n", keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:], pubkey)
		if !unique || !bytes.Equal(pubkey, keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
			return false, errors.Errorf("Crowdsale request must send tokens to DCB address")
		}
	}

	return true, nil
}

func (csReq *CrowdsaleRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(csReq.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("Wrong request info's payment address")
	}
	return false, true, nil
}

func (csReq *CrowdsaleRequest) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	// TODO(@0xbunyip): accept only some pairs of assets
	return true
}

func (csReq *CrowdsaleRequest) Hash() *common.Hash {
	record := csReq.PaymentAddress.String()
	record += string(csReq.SaleID)
	record += string(csReq.PriceLimit)
	record += string(csReq.ValidUntil)
	record += strconv.FormatBool(csReq.LimitSellingAssetPrice)

	// final hash
	record += csReq.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (csReq *CrowdsaleRequest) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	// Calculate value of asset sent in request tx
	saleData, err := bcr.GetCrowdsaleData(csReq.SaleID)
	if err != nil {
		return [][]string{}, err
	}
	sentAmount := uint64(0)
	if saleData.BuyingAsset.IsEqual(&common.ConstantID) {
		_, _, sentAmount = txr.GetUniqueReceiver()
	} else if common.IsBondAsset(&saleData.BuyingAsset) {
		_, _, sentAmount = txr.GetTokenUniqueReceiver()
	}
	lrActionValue := getCrowdsaleRequestActionValue(
		csReq.SaleID,
		csReq.PriceLimit,
		csReq.LimitSellingAssetPrice,
		csReq.PaymentAddress,
		sentAmount,
	)
	lrAction := []string{strconv.Itoa(CrowdsaleRequestMeta), lrActionValue}
	// TODO(@0xbunyip): BuildReqActions should return []string only?
	return [][]string{lrAction}, nil
}

func getCrowdsaleRequestActionValue(
	saleID []byte,
	priceLimit uint64,
	limitSell bool,
	paymentAddress privacy.PaymentAddress,
	sentAmount uint64,
) string {
	return strings.Join([]string{
		base64.StdEncoding.EncodeToString(saleID),
		strconv.FormatUint(priceLimit, 10),
		strconv.FormatBool(limitSell),
		paymentAddress.String(),
		strconv.FormatUint(sentAmount, 10),
	}, actionValueSep)
}

func ParseCrowdsaleRequestActionValue(values string) ([]byte, uint64, bool, privacy.PaymentAddress, uint64, error) {
	s := strings.Split(values, actionValueSep)
	if len(s) != 5 {
		return nil, 0, false, privacy.PaymentAddress{}, 0, errors.Errorf("CrowdsaleRequest value invalid")
	}
	saleID, errID := base64.StdEncoding.DecodeString(s[0])
	priceLimit, errPrice := strconv.ParseUint(s[1], 10, 64)
	limitSell, errSell := strconv.ParseBool(s[2])
	paymentAddressBytes, errPay := hex.DecodeString(s[3])
	sentAmount, errAmount := strconv.ParseUint(s[4], 10, 64)
	errSaver := &ErrorSaver{}
	if errSaver.Save(errID, errPrice, errSell, errPay, errAmount) != nil {
		return nil, 0, false, privacy.PaymentAddress{}, 0, errSaver.Get()
	}
	paymentAddress := privacy.NewPaymentAddressFromByte(paymentAddressBytes)
	return saleID, priceLimit, limitSell, *paymentAddress, sentAmount, nil
}

func (csReq *CrowdsaleRequest) CalculateSize() uint64 {
	return calculateSize(csReq)
}
