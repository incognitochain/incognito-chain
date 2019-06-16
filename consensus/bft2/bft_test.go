package bft2

import (
	"fmt"
	"github.com/constant-money/constant-chain/wire"
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
	"sync"
	"testing"
	"time"
)

type TestFrameWork struct {
	nodeList []*BFTEngine
}

type Block struct {
	Timestamp uint64
	Height uint64
	From string
	Round uint64
}

func (s Block) GetHeight() uint64 {
	return s.Height
}

func (s Block) GetRound() uint64 {
	return s.Round
}

func (s Block) GetProducer() string {
	return s.From
}

func (s Block) Hash() string {
	return fmt.Sprint(s.From,"_",s.Height,"_",s.Timestamp)
}

var Committees = []string{
	"17yV5NTyFPm73sHa7tK3mdKbMJmVavdrvZzhCynAexcg81BYfQe",
	"15gjFF5JCqTUFn2PSK4Bq7rQXBnU5qxRQQTfjjf3cZxGVd4ZQru",
	"17tN32moCx4QdmV9n9suPxqSCvrQPnujVNgvjwvFrKPTUU4r6rj",
	"14tcxUrXtR7daLG2LupmM28gjqL3FaAPRh5PDhxWJt6RkgYv1Ln",
	"18kbw2JgZxS6CdUj8yv47meXs3GRtcDhaa9FVbvx4U7EBXE4Ssi",
	"18Q9Hgc7k5qmeMnozJwiWM1hMKU5sRmfBkWawADuWkbpXzSmLcj",
	"16k96grxsDGG11N6iYKKuN7CVDSSiKZs2gzVjTMtHrRTmjbsZtk",
	"15e7jPK7PecXfTqhoL5uns8fPpCiniuUGp1TjiW4BDTdz6fxfxo",
	"16B6ox9Fd3T4hLegauBh68SdfvaQjZdL4tob65D1ku73uosJDJ5",
	"15PjLVkFtuD9DG1GcHcvLB34PB8GFDNR7nBQFvjWmKL9Rg8AURq",
}

type Chain struct {
	Block      []Block
	Committees []string
	Pubkey     string
	Env        TestFrameWork
}

func (s *Chain) PushMessageToValidator(m wire.Message) error {
	for _, node := range s.Env.nodeList {
		switch v := m.(type) {
		case *ProposeMsg:
			if node.Chain.GetNodePubKey() == s.GetNodePubKey(){
				continue
			}
			go func(node *BFTEngine){
				rand := common.RandInt() % 30000
				time.Sleep(time.Duration(rand)*time.Millisecond)
				node.ProposeMsgCh <- *m.(*ProposeMsg)
			}(node)
			
		case *PrepareMsg:
			if node.Chain.GetNodePubKey() == s.GetNodePubKey(){
				continue
			}
			go func(node *BFTEngine){
				rand := common.RandInt() % 6000
				time.Sleep(time.Duration(rand)*time.Millisecond)
				node.PrepareMsgCh <- *m.(*PrepareMsg)
			}(node)
			
		case *CommitMsg:
			if node.Chain.GetNodePubKey() == s.GetNodePubKey(){
				continue
			}
			go func(node *BFTEngine){
				rand := common.RandInt() % 6000
				time.Sleep(time.Duration(rand)*time.Millisecond)
				node.CommitMsgCh <- *m.(*CommitMsg)
			}(node)
			
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}
	}
	return nil
}

func (s *Chain) GetLastBlockTimeStamp() uint64 {
	return s.Block[len(s.Block)-1].Timestamp
}

func (s *Chain) GetBlkMinTime() time.Duration {
	return time.Second * 5
}

func (s *Chain) IsReady() bool {
	var maxHeight = uint64(0)
	for _, node := range s.Env.nodeList {
		if node.Chain.GetHeight() > maxHeight {
			maxHeight = node.Chain.GetHeight()
		}
	}
	if s.GetHeight() == maxHeight {
		return true
	}
	return false
}

func (s *Chain) GetHeight() uint64 {
	return s.Block[len(s.Block)-1].Height
}

func (s *Chain) GetLastBlock() Block{
	return s.Block[len(s.Block)-1]
}

func (s *Chain) GetNodePubKeyIndex() int {
	for i, v := range s.Committees {
		if v == s.Pubkey {
			return i
		}
	}
	return -1
}

func (s *Chain) GetCommitteeSize() int {
	return len(s.Committees)
}

func (s *Chain) GetNodePubKey() string {
	return s.Pubkey
}

func (s *Chain) CreateNewBlock() BlockInterface {
	return Block{Timestamp: uint64(time.Now().Unix()), Height: s.GetHeight() + 1, From: s.Pubkey}
}

func (s *Chain) ValidateBlock(blk interface{}) bool{
	blkData := blk.(Block)
	if s.GetHeight()+1 == blkData.GetHeight() {
		return true
	} else {
		return false
	}
}

func (s *Chain) ValidateSignature(blk interface{}, sig string) bool{
	return true
}

func (s *Chain) GetLastProposerIndex() int{
	return int(common.IndexOfStr(s.GetLastBlock().From, s.Committees))
}

func (s *Chain) InsertBlk(blk interface{}, willCommit bool) {
	blkData := blk.(Block)
	if s.GetHeight() >= blkData.Height {
		return
	}
	s.Block = append(s.Block, blkData)
	//fmt.Println(s.Block)
}

var NODE_NUM = 100
var testFramework = TestFrameWork{nodeList: make([]*BFTEngine, NODE_NUM)}

func TestBFTEngine_Start(t *testing.T) {
	for i := 0; i < NODE_NUM; i++ {
		newNode := new(BFTEngine)
		newNode.Chain = &Chain{Block: []Block{{1560675379, 1,"Genesis",0}}, Committees: Committees[:NODE_NUM], Pubkey: Committees[i], Env: testFramework}
		newNode.PeerID = strconv.Itoa(i)
		testFramework.nodeList[i] = newNode
		newNode.Start()
	}
	
	//go func(){
	//	ticker := time.Tick(time.Second*5)
	//	for _ = range ticker {
	//		for _,v := range testFramework.nodeList {
	//			v.debug(v.State,v.NextHeight, v.Round)
	//		}
	//	}
	//}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
