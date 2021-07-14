package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type PDEContributionV3 struct {
	PoolPairID          string // only "" for the first contribution of pool
	PairHash            string
	OtaPublicKeyRefund  string // refund contributed token
	OtaTxRandomRefund   string
	OtaPublicKeyReceive string // receive nfct
	OtaTxRandomReceive  string
	TokenID             string
	TokenAmount         uint64
	Amplifier           uint // only set for the first contribution
	MetadataBase
}

func NewPDEContributionV3() *PDEContributionV3 {
	return &PDEContributionV3{}
}

func NewPDEContributionV3WithValue(
	poolPairID, pairHash,
	otaPublicKeyRefund, otaTxRandomRefund,
	otaPublicKeyReceive, otaTxRandomReceive,
	tokenID string, tokenAmount uint64, amplifier uint,
) *PDEContributionV3 {
	return &PDEContributionV3{
		PoolPairID:          poolPairID,
		PairHash:            pairHash,
		OtaPublicKeyRefund:  otaPublicKeyRefund,
		OtaTxRandomRefund:   otaTxRandomRefund,
		OtaPublicKeyReceive: otaPublicKeyReceive,
		OtaTxRandomReceive:  otaTxRandomReceive,
		TokenID:             tokenID,
		TokenAmount:         tokenAmount,
		Amplifier:           amplifier,
	}
}

func (pc *PDEContributionV3) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (pc *PDEContributionV3) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	/*if chainRetriever.IsAfterPrivacyV2CheckPoint(beaconHeight) && pc.GetType() == PDEContributionMeta {*/
	//return false, false, fmt.Errorf("metadata type %v is no longer supported, consider using %v instead", PDEContributionMeta, PDEPRVRequiredContributionRequestMeta)
	//}

	//if pc.PDEContributionPairID == "" {
	//return false, false, errors.New("PDE contribution pair id should not be empty.")
	//}

	//if _, err := AssertPaymentAddressAndTxVersion(pc.ContributorAddressStr, tx.GetVersion()); err != nil {
	//return false, false, err
	//}

	//isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	//if err != nil || !isBurned {
	//return false, false, errors.New("Error This is not Tx Burn")
	//}

	//if pc.ContributedAmount == 0 || pc.ContributedAmount != burnCoin.GetValue() {
	//return false, false, errors.New("Contributed Amount is not valid ")
	//}

	//tokenID, err := common.Hash{}.NewHashFromStr(pc.TokenIDStr)
	//if err != nil {
	//return false, false, NewMetadataTxError(IssuingRequestNewIssuingRequestFromMapEror, errors.New("TokenIDStr incorrect"))
	//}
	//if !bytes.Equal(burnedTokenID[:], tokenID[:]) {
	//return false, false, errors.New("Wrong request info's token id, it should be equal to tx's token id.")
	//}

	//if tx.GetType() == common.TxNormalType && pc.TokenIDStr != common.PRVCoinID.String() {
	//return false, false, errors.New("With tx normal privacy, the tokenIDStr should be PRV, not custom token.")
	//}

	//if tx.GetType() == common.TxCustomTokenPrivacyType && pc.TokenIDStr == common.PRVCoinID.String() {
	//return false, false, errors.New("With tx custome token privacy, the tokenIDStr should not be PRV, but custom token.")
	/*}*/

	return true, true, nil
}

func (pc *PDEContributionV3) ValidateMetadataByItself() bool {
	return pc.Type == PDexV3AddLiquidityMeta
}

func (pc *PDEContributionV3) Hash() *common.Hash {
	/*record := pc.MetadataBase.Hash().String()*/
	//record += pc.PDEContributionPairID
	//record += pc.ContributorAddressStr
	//record += pc.TokenIDStr
	//record += strconv.FormatUint(pc.ContributedAmount, 10)
	//// final hash
	//hash := common.HashH([]byte(record))
	/*return &hash*/
	return nil
}

func (pc *PDEContributionV3) BuildReqActions(
	tx Transaction,
	chainRetriever ChainRetriever,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	/*actionContent := PDEContributionAction{*/
	//Meta:    *pc,
	//TxReqID: *tx.Hash(),
	//ShardID: shardID,
	//}
	//actionContentBytes, err := json.Marshal(actionContent)
	//if err != nil {
	//return [][]string{}, err
	//}
	//actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	//action := []string{strconv.Itoa(pc.Type), actionContentBase64Str}
	/*return [][]string{action}, nil*/
	return [][]string{}, nil
}

func (pc *PDEContributionV3) CalculateSize() uint64 {
	//return calculateSize(pc)
	return 0
}
