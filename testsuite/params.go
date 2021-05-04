package devframework

import "github.com/incognitochain/incognito-chain/blockchain"

const (
	ID_MAINNET = iota
	ID_TESTNET
	ID_TESTNET2
	ID_CUSTOM
)

type ChainParam struct {
	blockchain.Params
}

func (s *ChainParam) SetActiveShardNumber(nShard int) *ChainParam {
	s.ActiveShards = nShard
	return s
}
func (s *ChainParam) SetMaxShardCommitteeSize(maxNum int) *ChainParam {
	s.MaxShardCommitteeSize = maxNum
	return s
}
func (s *ChainParam) SetBCHeightBreakPointNewZKP(height uint64) *ChainParam {
	s.BCHeightBreakPointNewZKP = height
	return s
}
func (s *ChainParam) GetParamData() *blockchain.Params {
	return &s.Params
}

func NewChainParam(networkID int) *ChainParam {
	//setup param
	blockchain.SetupParam()
	switch networkID {
	case ID_MAINNET:
		return &ChainParam{blockchain.ChainMainParam}
	case ID_TESTNET:
		return &ChainParam{blockchain.ChainTestParam}
	case ID_TESTNET2:
		return &ChainParam{blockchain.ChainTest2Param}
	}
	return &ChainParam{blockchain.ChainTest2Param}
}
