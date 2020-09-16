package main

import consensus "github.com/incognitochain/incognito-chain/consensus_v2"

type FakeDB struct{
	genesisBlock consensus.BlockInterface
}

func (s *FakeDB) GetGenesisBlock() consensus.BlockInterface {
	return s.genesisBlock
}
