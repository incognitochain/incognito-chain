package metadata

// import "github.com/ninjadotorg/constant/common"

// type BuyBackRequest struct {
// 	BuyBackFromTxID common.Hash
// 	VoutIndex       int
// }

// func NewBuyBackRequest(bbReqData map[string]interface{}) *BuyBackRequest {
// 	return &BuyBackRequest{
// 		BuyBackFromTxID: bbReqData["buyBackFromTxId"].(common.Hash),
// 		VoutIndex:       bbReqData["voutIndex"].(int),
// 	}
// }

// func (bbReq *BuyBackRequest) ValidateTxWithBlockChain(
// 	txr Transaction,
// 	bcr BlockchainRetriever,
// 	chainID byte,
// ) (bool, error) {

// }

// func (bbReq *BuyBackRequest) ValidateSanityData(
// 	txr Transaction,
// ) (bool, bool, error) {

// }

// func (bsReq *BuyBackRequest) ValidateMetadataByItself() bool {
// 	// The validation just need to check at tx level, so returning true here
// 	return true
// }

// func (bsReq *BuyBackRequest) Hash() *common.Hash {
// 	record := string(bsReq.PaymentAddress.ToBytes())
// 	record += bsReq.AssetType.String()
// 	record += string(bsReq.Amount)
// 	record += string(bsReq.BuyPrice)
// 	record += string(bsReq.SaleID)

// 	// final hash
// 	hash := common.DoubleHashH([]byte(record))
// 	return &hash
// }
