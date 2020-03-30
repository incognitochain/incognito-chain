package lvdb

//// prefix key for portal
//var (
//	CustodianStatePrefix                = []byte("custodianstate-")
//	PortalPortingRequestsPrefix         = []byte("portalportingrequest-")
//	PortalPortingRequestsTxPrefix       = []byte("portalportingrequesttx-")
//	PortalExchangeRatesPrefix           = []byte("portalexchangeratesrequest-")
//	PortalFinalExchangeRatesPrefix      = []byte("portalfinalexchangerates-")
//	PortalCustodianStatePrefix          = []byte("portalcustodianstate-")
//	PortalCustodianDepositPrefix        = []byte("portalcustodiandeposit-")
//	PortalWaitingPortingRequestsPrefix  = []byte("portalwaitingportingrequest-")
//	PortalRequestPTokensPrefix          = []byte("portalrequestptokens-")
//	PortalWaitingRedeemRequestsPrefix   = []byte("portalwaitingredeemrequest-")
//	PortalRedeemRequestsPrefix          = []byte("portalredeemrequest-")
//	PortalRedeemRequestsByTxReqIDPrefix = []byte("portalredeemrequestbytxid-")
//	PortalRequestUnlockCollateralPrefix = []byte("portalrequestunlockcollateral-")
//	PortalCustodianWithdrawPrefix       = []byte("portalcustodianwithdraw-")
//
//	// liquidation in portal
//	PortalLiquidateCustodianPrefix                  = []byte("portalliquidatecustodian-")
//	PortalLiquidateTopPercentileExchangeRatesPrefix = []byte("portalliquidatetoppercentileexchangerates-")
//	PortalLiquidateExchangeRatesPrefix              = []byte("portalliquidateexchangerates-")
//	PortalLiquidationCustodianDepositPrefix         = []byte("portalliquidationcustodiandepsit-")
//
//	PortalExpiredPortingReqPrefix = []byte("portalexpiredportingreq-")
//
//	// reward in portal
//	PortalRewardByBeaconHeightPrefix  = []byte("portalreward-")
//	PortalRequestWithdrawRewardPrefix = []byte("portalrequestwithdrawreward-")
//
//	// Relaying
//	RelayingBNBHeaderStatePrefix = []byte("relayingbnbheaderstate-")
//	RelayingBNBHeaderChainPrefix = []byte("relayingbnbheaderchain-")
//)


//type RemoteAddress struct {
//	PTokenID string
//	Address  string
//}
//
//type CustodianState struct {
//	IncognitoAddress       string
//	TotalCollateral        uint64            // prv
//	FreeCollateral         uint64            // prv
//	HoldingPubTokens       map[string]uint64 // tokenID : amount
//	LockedAmountCollateral map[string]uint64 // tokenID : amount
//	RemoteAddresses        []RemoteAddress   // tokenID : address
//	RewardAmount           uint64            // reward in prv
//}
//
//type MatchingPortingCustodianDetail struct {
//	IncAddress             string
//	RemoteAddress          string
//	Amount                 uint64
//	LockedAmountCollateral uint64
//	RemainCollateral       uint64
//}
//
//type MatchingRedeemCustodianDetail struct {
//	IncAddress    string
//	RemoteAddress string
//	Amount        uint64
//}
//
//type PortingRequest struct {
//	UniquePortingID string
//	TxReqID         common.Hash
//	TokenID         string
//	PorterAddress   string
//	Amount          uint64
//	Custodians      []*MatchingPortingCustodianDetail
//	PortingFee      uint64
//	Status          int
//	BeaconHeight    uint64
//}
//
//type RedeemRequest struct {
//	UniqueRedeemID        string
//	TxReqID               common.Hash
//	TokenID               string
//	RedeemerAddress       string
//	RedeemerRemoteAddress string
//	RedeemAmount          uint64
//	Custodians            []*MatchingRedeemCustodianDetail
//	RedeemFee             uint64
//	BeaconHeight          uint64
//}
//
//type ExchangeRatesRequest struct {
//	SenderAddress string
//	Rates         []*ExchangeRateInfo
//}
//
//type FinalExchangeRatesDetail struct {
//	Amount uint64
//}
//
//type FinalExchangeRates struct {
//	Rates map[string]FinalExchangeRatesDetail
//}
//
//type CustodianWithdrawRequest struct {
//	PaymentAddress                string
//	Amount                        uint64
//	Status                        int
//	RemainCustodianFreeCollateral uint64
//}
//
//type LiquidateTopPercentileExchangeRatesDetail struct {
//	TPKey int
//	TPValue                  int
//	HoldAmountFreeCollateral uint64
//	HoldAmountPubToken       uint64
//}
//
//type LiquidateTopPercentileExchangeRates struct {
//	CustodianAddress 	string
//	Status				byte
//	Rates        		map[string]LiquidateTopPercentileExchangeRatesDetail //ptoken | detail
//}
//
//type LiquidateExchangeRatesDetail struct {
//	HoldAmountFreeCollateral uint64
//	HoldAmountPubToken       uint64
//}
//
//type LiquidateExchangeRates struct {
//	Rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
//}
//
//type RedeemLiquidateExchangeRates struct {
//	TxReqID               common.Hash
//	TokenID               string
//	RedeemerAddress       string
//	RedeemerRemoteAddress string
//	RedeemAmount          uint64
//	RedeemFee             uint64
//	Status                byte
//	TotalPTokenReceived	  uint64
//}
//
//type LiquidationCustodianDeposit struct {
//	TxReqID common.Hash
//	IncogAddressStr string
//	PTokenId string
//	DepositAmount uint64
//	FreeCollateralSelected bool
//	Status byte
//}
//
//func NewCustodianWithdrawRequestKey(txHash string) string {
//	key := append(PortalCustodianWithdrawPrefix, []byte(txHash)...)
//	return string(key)
//}
//
//func NewCustodianStateKey(beaconHeight uint64, custodianAddress string) string {
//	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
//	key := append(PortalCustodianStatePrefix, beaconHeightBytes...)
//	key = append(key, []byte(custodianAddress)...)
//	return string(key)
//}
//
//func NewPortingRequestKey(uniquePortingID string) string {
//	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
//	return string(key) //prefix + uniqueId
//}
//
//func NewPortingRequestTxKey(txReqID string) string {
//	key := append(PortalPortingRequestsTxPrefix, []byte(txReqID)...)
//	return string(key) //prefix + txHash
//}
//
//func NewFinalExchangeRatesKey(beaconHeight uint64) string {
//	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
//	key := append(PortalFinalExchangeRatesPrefix, beaconHeightBytes...)
//	key = append(key, []byte("portal")...)
//	return string(key)
//}
//
//func NewExchangeRatesRequestKey(beaconHeight uint64, txId string) string {
//	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
//	key := append(PortalExchangeRatesPrefix, beaconHeightBytes...)
//	key = append(key, []byte(txId)...)
//	return string(key)
//}
//
//func NewCustodianDepositKey(txID string) string {
//	key := append(PortalCustodianDepositPrefix, []byte(txID)...)
//	return string(key)
//}
//
//func NewWaitingPortingReqKey(beaconHeight uint64, portingID string) string {
//	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
//	key := append(PortalWaitingPortingRequestsPrefix, beaconHeightBytes...)
//	key = append(key, []byte(portingID)...)
//	return string(key)
//}
//
//// NewPortalReqPTokenKey creates key for tracking request pToken in portal
//func NewPortalReqPTokenKey(txReqStr string) string {
//	key := append(PortalRequestPTokensPrefix, []byte(txReqStr)...)
//	return string(key)
//}
//
//// GetRemoteAddressByTokenID returns remote address for tokenID
//func GetRemoteAddressByTokenID(addresses []RemoteAddress, tokenID string) (string, error) {
//	for _, addr := range addresses {
//		if addr.PTokenID == tokenID {
//			return addr.Address, nil
//		}
//	}
//
//	return "", errors.New("Can not found address with tokenID")
//}
//
//type PortalRewardInfo struct {
//	CustodianIncAddr string
//	Amount           uint64
//}
//
//type ExchangeRateInfo struct {
//	PTokenID string
//	Rate     uint64
//}

//// TrackCustodianDepositCollateral stores custodian deposit collateral into db with deposit TxID
//func (db *db) TrackCustodianDepositCollateral(key []byte, content []byte) error {
//	err := db.Put(key, content)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// GetCustodianDepositCollateralStatus returns custodian deposit status with deposit TxID
//func (db *db) GetCustodianDepositCollateralStatus(txIDStr string) ([]byte, error) {
//	key := NewCustodianDepositKey(txIDStr)
//	custodianDepositStatusBytes, err := db.lvdb.Get([]byte(key), nil)
//	if err != nil && err != lvdberr.ErrNotFound {
//		return nil, database.NewDatabaseError(database.GetCustodianDepositStatusError, err)
//	}
//
//	return custodianDepositStatusBytes, err
//}
//
//// TrackReqPTokens stores request ptokens into db with req TxID
//func (db *db) TrackReqPTokens(key []byte, content []byte) error {
//	err := db.Put(key, content)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// GetReqPTokenStatusByTxReqID returns request ptoken status with txReqID
//func (db *db) GetReqPTokenStatusByTxReqID(txReqID string) ([]byte, error) {
//	key := append(PortalRequestPTokensPrefix, []byte(txReqID)...)
//
//	reqPTokenStatusBytes, err := db.lvdb.Get(key, nil)
//	if err != nil && err != lvdberr.ErrNotFound {
//		return nil, database.NewDatabaseError(database.GetReqPTokenStatusError, err)
//	}
//
//	return reqPTokenStatusBytes, err
//}
//
//// ======= REDEEM =======
//// NewWaitingRedeemReqKey creates key for storing waiting redeems list in portal
////func NewWaitingRedeemReqKey(beaconHeight uint64, redeemID string) string {
////	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
////	key := append(PortalWaitingRedeemRequestsPrefix, beaconHeightBytes...)
////	key = append(key, []byte(redeemID)...)
////	return string(key)
////}
////
////// NewRedeemReqKey creates key for tracking redeems status in portal
////func NewRedeemReqKey(redeemID string) string {
////	key := append(PortalRedeemRequestsPrefix, []byte(redeemID)...)
////	return string(key)
////}
////
////// NewRedeemReqKey creates key for tracking redeems status in portal
////func NewTrackRedeemReqByTxReqIDKey(txID string) string {
////	key := append(PortalRedeemRequestsByTxReqIDPrefix, []byte(txID)...)
////	return string(key)
////}
//
//// StoreRedeemRequest stores status of redeem request by redeemID
//func (db *db) StoreRedeemRequest(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.StoreRedeemRequestError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// GetRedeemRequestByRedeemID returns status of redeem request by redeemID
//func (db *db) GetRedeemRequestByRedeemID(redeemID string) ([]byte, error) {
//	key := NewRedeemReqKey(redeemID)
//	return db.GetItemPortalByKey([]byte(key))
//}
//
//// TrackRedeemRequestByTxReqID tracks status of redeem request by txReqID
//func (db *db) TrackRedeemRequestByTxReqID(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackRedeemReqByTxReqIDError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//
//// NewPortalReqUnlockCollateralKey creates key for tracking request unlock collateral in portal
////func NewPortalReqUnlockCollateralKey(txReqStr string) string {
////	key := append(PortalRequestUnlockCollateralPrefix, []byte(txReqStr)...)
////	return string(key)
////}
//
//// TrackRequestUnlockCollateralByTxReqID tracks status of request unlock collateral by txReqID
//func (db *db) TrackRequestUnlockCollateralByTxReqID(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackRedeemReqByTxReqIDError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// GetReqUnlockCollateralStatusByTxReqID returns request unlock collateral status with txReqID
//func (db *db) GetReqUnlockCollateralStatusByTxReqID(txReqID string) ([]byte, error) {
//	key := NewPortalReqUnlockCollateralKey(txReqID)
//
//	reqUnlockCollateralStatusBytes, err := db.lvdb.Get([]byte(key), nil)
//	if err != nil && err != lvdberr.ErrNotFound {
//		return nil, database.NewDatabaseError(database.GetReqUnlockCollateralStatusError, err)
//	}
//
//	return reqUnlockCollateralStatusBytes, err
//}
//
//// NewPortalLiquidationCustodianKey creates key for tracking custodian liquidation in portal
////func NewPortalLiquidationCustodianKey(redeemID string, custodianIncAddrStr string) string {
////	key := append(PortalLiquidateCustodianPrefix, []byte(redeemID)...)
////	key = append(key, []byte(custodianIncAddrStr)...)
////	return string(key)
////}
////
////// NewPortalLiquidationCustodianKey creates key for tracking custodian liquidation in portal
////func NewPortalExpiredPortingReqKey(portingID string) string {
////	key := append(PortalExpiredPortingReqPrefix, []byte(portingID)...)
////	return string(key)
////}
//
//// TrackLiquidateCustodian tracks status of liquidation custodian by txReqID
//func (db *db) TrackLiquidateCustodian(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackLiquidateCustodianError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// TrackRequestUnlockCollateralByTxReqID tracks status of request unlock collateral by txReqID
//func (db *db) GetLiquidateCustodian(redeemID string, custodianIncAddrStr string) ([]byte, error) {
//	key := NewPortalLiquidationCustodianKey(redeemID, custodianIncAddrStr)
//	return db.GetItemPortalByKey([]byte(key))
//}
//
////func NewPortalLiquidateTPExchangeRatesKey(beaconHeight uint64, custodianIncAddrStr string) string {
////	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
////	key := append(PortalLiquidateTopPercentileExchangeRatesPrefix, beaconHeightBytes...)
////	key = append(key, []byte(custodianIncAddrStr)...)
////	return string(key)
////}
//
//// TrackExpiredPortingReq tracks status of expired waiting porting request by PortingID
//func (db *db) TrackExpiredPortingReq(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.TrackLiquidateCustodianError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
////func NewPortalLiquidateExchangeRatesKey(beaconHeight uint64) string {
////	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
////	key := append(PortalLiquidateExchangeRatesPrefix, beaconHeightBytes...)
////	key = append(key, []byte("liquidation")...)
////	return string(key)
////}
//
//
//// NewPortalRewardKey creates key for storing portal reward by beacon height
////func NewPortalRewardKey(beaconHeight uint64) string {
////	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
////	key := append(PortalRewardByBeaconHeightPrefix, beaconHeightBytes...)
////	return string(key)
////}
//
//// StorePortalRewardByBeaconHeight stores portal reward by beacon height
//func (db *db) StorePortalRewardByBeaconHeight(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.StorePortalRewardError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
//// NewPortalReqWithdrawRewardKey creates key for storing request withdraw portal reward
////func NewPortalReqWithdrawRewardKey(beaconHeight uint64, custodianAddr string) string {
////	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
////	key := append(PortalRequestWithdrawRewardPrefix, beaconHeightBytes...)
////	key = append(key, []byte(custodianAddr)...)
////	return string(key)
////}
//
//// TrackPortalReqWithdrawReward stores portal request withdraw portal reward
//func (db *db) TrackPortalReqWithdrawReward(key []byte, value []byte) error {
//	err := db.Put(key, value)
//	if err != nil {
//		return database.NewDatabaseError(database.StorePortalRewardError, errors.Wrap(err, "db.lvdb.put"))
//	}
//	return nil
//}
//
////func NewRedeemLiquidateExchangeRatesKey(txId string) string {
////	key := append(PortalWaitingRedeemRequestsPrefix, []byte(txId)...)
////	return string(key)
////}
//
//
////func NewLiquidationCustodianDepositKey(txID string) string  {
////	key := append(PortalCustodianDepositPrefix, []byte(txID)...)
////	return string(key)
////}
//
