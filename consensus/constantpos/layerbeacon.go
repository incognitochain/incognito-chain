package constantpos

type Layerbeacon struct {
	cQuit                 chan struct{}
	Committee             CommitteeStruct
	started               bool
	protocol              *BFTProtocol
	knownChainsHeight     chainsHeight
	validatedChainsHeight chainsHeight
}

func (self *Layerbeacon) Start() {

}

func (self *Layerbeacon) Stop() {

}

func (self *Layerbeacon) BeaconWatcher() {

}
