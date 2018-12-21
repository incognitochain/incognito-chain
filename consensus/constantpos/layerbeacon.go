package constantpos

type Layerbeacon struct {
	cQuit     chan struct{}
	Committee CommitteeStruct
	started   bool
	protocol  *BFTProtocol
}

func (self *Layerbeacon) Start() {

}

func (self *Layerbeacon) Stop() {

}

func (self *Layerbeacon) BeaconWatcher() {

}
