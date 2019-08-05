package main

import (
	"testing"
)

func TestGetBalanceByPrivateKey(t *testing.T) {
	c := newClientWithFullInform("http://localhost", "9334","19334")
	res, err := c.getBalanceByPrivatekey("112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
func TestGetTransactionByHash(t *testing.T) {
	c := newClientWithFullInform("http://localhost", "9334","19334")
	res, err := c.getTransactionByHash("3488cca3c7cc6d29ddc1a739135cc9f269f7c47e8177a0f735e47072a9a71c64")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
func TestGetBlockChainInfo(t *testing.T) {
	c := newClientWithFullInform("http://localhost", "9334","19334")
	res, err := c.getBlockChainInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
func TestCreateAndSendTransaction(t *testing.T) {
	c := newClientWithFullInform("http://localhost", "9334","19334")
	receiver := make(map[string]uint64)
	receiver["1Uv2wgU5FR5jjeN3uY3UJ4SYYyjqj97spYBEDa6cTLGiP3w6BCY7mqmASKwXz8hXfLr6mpDjhWDJ8TiM5v5U5f2cxxqCn5kwy5JM9wBgi"] = 50
	params := []interface{}{
		"112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe",
		receiver,
		15,
		0,
	}
	res, err := c.createAndSendTransaction(params)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
