package rpcserver

import (
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/wallet"
)

func (self RpcServer) handleGetAmountVoteToken(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	paymentAddress := arrayParams[0].(string)
	pubKey := wallet.GetPubKeyFromPaymentAddress(paymentAddress)
	db := *self.config.Database
	result := jsonresult.ListCustomTokenBalance{ListCustomTokenBalance: []jsonresult.CustomTokenBalance{}}

	// For DCB voting token
	result.PaymentAddress = paymentAddress
	item := jsonresult.CustomTokenBalance{}
	item.Name = "DCB voting token"
	item.Symbol = "DCB Voting Token"
	TokenID := &common.Hash{}
	TokenID.SetBytes(common.DCBVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err := db.GetVoteTokenAmount("dcb", self.config.BlockChain.GetCurrentBoardIndex(blockchain.DCBConstitutionHelper{}), pubKey)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	// For GOV voting token
	item = jsonresult.CustomTokenBalance{}
	item.Name = "GOV voting token"
	item.Symbol = "GOV Voting Token"
	TokenID = &common.Hash{}
	TokenID.SetBytes(common.GOVVotingTokenID[:])
	item.TokenID = TokenID.String()
	item.TokenImage = common.Render([]byte(item.TokenID))
	amount, err = db.GetVoteTokenAmount("gov", self.config.BlockChain.GetCurrentBoardIndex(blockchain.GOVConstitutionHelper{}), pubKey)
	if err != nil {
		Logger.log.Error(err)
	}
	item.Amount = uint64(amount)
	result.ListCustomTokenBalance = append(result.ListCustomTokenBalance, item)

	return result, nil
}

func (self RpcServer) handleGetEncryptionFlag(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {

}
