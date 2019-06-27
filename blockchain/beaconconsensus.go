package blockchain

import (
	"github.com/incognitochain/incognito-chain/consensus/bft"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

func (beaconChain *BestStateBeacon) PushMessageToValidator(wire.Message) error {

	return nil
}

func (beaconChain *BestStateBeacon) GetLastBlockTimeStamp() uint64 {
	return 0
}

func (beaconChain *BestStateBeacon) GetBlkMinTime() time.Duration {
	return time.Second * 5

}

func (beaconChain *BestStateBeacon) IsReady() bool {
	return true
}

func (beaconChain *BestStateBeacon) GetHeight() uint64 {
	return 1
}

func (beaconChain *BestStateBeacon) GetCommitteeSize() int {
	return 0
}

func (beaconChain *BestStateBeacon) GetNodePubKeyIndex() int {
	return 0
}

func (beaconChain *BestStateBeacon) GetLastProposerIndex() int {
	return 0
}

func (beaconChain *BestStateBeacon) GetNodePubKey() string {
	return "0"
}

func (beaconChain *BestStateBeacon) CreateNewBlock() bft.BlockInterface {
	return nil
}

func (beaconChain *BestStateBeacon) ValidateBlock(interface{}) bool {
	return true
}

func (beaconChain *BestStateBeacon) ValidateSignature(interface{}, string) bool {
	return true
}

func (beaconChain *BestStateBeacon) InsertBlk(interface{}, bool) {
	return
}
