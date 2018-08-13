package main

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/connmanager"
)

type Server struct {
	ChainParams *blockchain.Params
	ConnManager *connmanager.ConnManager

	Quit chan struct{}
}
