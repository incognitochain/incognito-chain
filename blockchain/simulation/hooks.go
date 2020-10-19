package main

import "github.com/incognitochain/incognito-chain/common"

func (sim *simInstance) SetPreCreateBlock(f func()) {
	sim.hooks.preCreateBlock = f
}

func (sim *simInstance) SetPostCreateBlock(f func(block common.BlockInterface)) {
	sim.hooks.postCreateBlock = f
}

func (sim *simInstance) SetPreInsertBlock(f func(block common.BlockInterface)) {
	sim.hooks.preInsertBlock = f
}

func (sim *simInstance) SetPostInsertBlock(f func()) {
	sim.hooks.postInsertBlock = f
}
