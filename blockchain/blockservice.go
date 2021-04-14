package blockchain

import (
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/metadata"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"github.com/pkg/errors"
)

func (bc *BlockChain) GetBlockByHeight(
	blkType proto.BlkType,
	height uint64,
	fromcID byte,
	tocID byte,
) (
	interface{},
	error,
) {
	switch blkType {
	case proto.BlkType_BlkBc:
		return bc.GetBeaconBlockByHeightV1(height)
	case proto.BlkType_BlkShard:
		return bc.GetShardBlockByHeightV1(height, fromcID)
	case proto.BlkType_BlkXShard:
		blk, err := bc.GetShardBlockByHeightV1(height, fromcID)
		if err != nil {
			return nil, err
		}
		return CreateCrossShardBlock(blk, tocID)
	default:
		return nil, errors.Errorf("Invalid block type")
	}
}

func (bc *BlockChain) GetBlockByHash(
	blkType proto.BlkType,
	hash *common.Hash,
	fromcID byte,
	tocID byte,
) (
	interface{},
	error,
) {
	switch blkType {
	case proto.BlkType_BlkBc:
		blk, _, err := bc.GetBeaconBlockByHash(*hash)
		if err != nil {
			return nil, err
		}
		return blk, nil
	case proto.BlkType_BlkShard:
		blk, _, err := bc.GetShardBlockByHash(*hash)
		if err != nil {
			return nil, err
		}
		return blk, nil
	case proto.BlkType_BlkXShard:
		blk, _, err := bc.GetShardBlockByHash(*hash)
		if err != nil {
			return nil, err
		}
		return CreateCrossShardBlock(blk, tocID)
	default:
		return nil, errors.Errorf("Invalid block type")
	}
}

func (bc *BlockChain) StreamBlockByHeight(
	fromPool bool,
	req *proto.BlockByHeightRequest,
) chan interface{} {
	// Logger.log.Infof("[stream] Netsync received request get block %v %v [%v...%v] len %v", fromPool, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights))
	Logger.log.Infof("[stream] Netsync received request stream block type %v, spec %v, height [%v..%v] len %v, from %v to %v uuid %v", req.Type, req.Specific, req.Heights[0], req.Heights[len(req.Heights)-1], len(req.Heights), req.From, req.To, req.UUID)
	blkCh := make(chan interface{})
	if !req.Specific {
		if len(req.Heights) != 2 || req.Heights[1] < req.Heights[0] {
			return nil
		}
	}
	sort.Slice(req.Heights, func(i, j int) bool { return req.Heights[i] < req.Heights[j] })
	go bc.streamBlkByHeight(req, blkCh)
	return blkCh
}

func (bc *BlockChain) streamBlkByHeight(
	req *proto.BlockByHeightRequest,
	blkCh chan interface{},
) {

	blkHeight := req.Heights[0] - 1
	idx := 0
	for blkHeight < req.Heights[len(req.Heights)-1] {
		if req.Specific {
			blkHeight = req.Heights[idx]
			idx++
		} else {
			blkHeight++
		}
		if blkHeight <= 1 {
			continue
		}
		blk, err := bc.GetBlockByHeight(req.Type, blkHeight, byte(req.From), byte(req.To))
		if err != nil {
			Logger.log.Errorf("[stream] Netsync cannot get block, return error %+v uuid %v", err, req.UUID)
			break
		}
		blkCh <- blk
	}
	close(blkCh)
	return
}

func (bc *BlockChain) StreamBlockByHash(
	fromPool bool,
	req *proto.BlockByHashRequest,
) chan interface{} {
	Logger.log.Infof("[stream] Netsync received request stream block type %v, hashes [%v..%v] len %v, from %v to %v uuid %v", req.Type, req.Hashes[0], req.Hashes[len(req.Hashes)-1], len(req.Hashes), req.From, req.To, req.UUID)
	blkCh := make(chan interface{})
	go bc.streamBlkByHash(req, blkCh)
	return blkCh
}

func (bc *BlockChain) streamBlkByHash(
	req *proto.BlockByHashRequest,
	blkCh chan interface{},
) {
	for _, blkHashByte := range req.GetHashes() {
		blkHash := &common.Hash{}
		blkHash.SetBytes(blkHashByte)
		blk, err := bc.GetBlockByHash(req.Type, blkHash, byte(req.From), byte(req.To))
		if err != nil {
			Logger.log.Errorf("[stream] Netsync cannot get block, return error %+v, uuid %v", err, req.UUID)
			break
		}
		blkCh <- blk
	}
	close(blkCh)
	return
}

func checkLimitTxAction(
	forBeacon bool,
	remining map[int]int,
	blk *types.ShardBlock,
) bool {
	txs := blk.Body.Transactions
	count := map[int]int{}
	for _, tx := range txs {
		act := metadata.GetMetaAction(tx.GetMetadataType())
		if act == metadata.NoAction {
			continue
		}
		if total, ok := count[act]; ok {
			count[act] = total + 1
		} else {
			count[act] = 1
		}
		if _, ok := remining[act]; !ok {
			remining[act] = metadata.GetLimitOfMeta(tx.GetMetadataType())
		}
		limit := remining[act]
		if limit-count[act] < 0 {
			if forBeacon {
				if act == metadata.MetaRequestBeaconMintTxs {
					return false
				}
			} else {
				return false
			}
		}
	}
	for k, v := range count {
		remining[k] = remining[k] - v
	}
	return true
}
