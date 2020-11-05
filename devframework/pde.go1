package devframework

import "github.com/incognitochain/incognito-chain/rpcserver/jsonresult"

func (sim *SimulationEngine) API_CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := sim.rpc_createandsendprivacycustomtokentransaction(privateKey, receivers, fee, privacy, tokenInfo)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := sim.rpc_createandsendtxwithwithdrawalreqv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := sim.rpc_createandsendtxwithpdefeewithdrawalreq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := sim.rpc_createandsendtxwithptokentradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := sim.rpc_createandsendtxwithptokencrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := sim.rpc_createandsendtxwithprvtradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := sim.rpc_createandsendtxwithprvcrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := sim.rpc_createandsendtxwithptokencontributionv2(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (sim *SimulationEngine) API_CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := sim.rpc_createandsendtxwithprvcontributionv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (sim *SimulationEngine) API_GetPDEState(beaconHeight float64) (interface{}, error) {
	result, err := sim.rpc_getpdestate(map[string]interface{}{"BeaconHeight": beaconHeight})
	return &result, err
}
