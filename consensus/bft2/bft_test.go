package bft2

import (
	"fmt"
	"github.com/constant-money/constant-chain/wire"
	"strconv"
	"sync"
	"testing"
	"time"
)

type TestFrameWork struct {
	nodeList []*BFTEngine
}

type Block struct {
	Height uint64
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
	Role       Role
	Committees []string
	Pubkey     string
	Env        TestFrameWork
}

func (s *Chain) PushMessageToValidator(m wire.Message) error {
	for _, node := range s.Env.nodeList {
		switch v := m.(type) {
		case *View:
			node.ViewMsgCh <- *(m.(*View))
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}
	}
	return nil
}

func (s *Chain) GetLastBlockTimeStamp() int64 {
	return 0
}

func (s *Chain) GetBlkMinTime() time.Duration {
	return time.Second * 5
}

func (s *Chain) IsReady() bool {
	return false
}

func (s *Chain) GetRole() Role {
	return s.Role
}

func (s *Chain) GetHeight() uint64 {
	return s.Block[len(s.Block)-1].Height
}

func (s *Chain) GetCommitteeSize() int {
	return len(s.Committees)
}

func (s *Chain) GetPubKey() string {
	return s.Pubkey
}

var NODE_NUM = 6
var testFramework = TestFrameWork{nodeList: make([]*BFTEngine, NODE_NUM)}

func TestBFTEngine_Start(t *testing.T) {
	for i := 0; i < NODE_NUM; i++ {
		newNode := new(BFTEngine)
		newNode.Chain = &Chain{Block: []Block{{1}}, Role: Role{"shard", "validator", 0}, Committees: Committees[:NODE_NUM], Pubkey: Committees[i], Env: testFramework}
		newNode.IsReady = false
		newNode.PeerID = strconv.Itoa(i)
		testFramework.nodeList[i] = newNode
		newNode.Start()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}
