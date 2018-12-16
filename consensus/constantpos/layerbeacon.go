package constantpos

type Layerbeacon struct {
	cQuit                 chan struct{}
	Committee             CommitteeStruct
	started               bool
	protocol              *BFTProtocol
	knownChainsHeight     int
	validatedChainsHeight int
}

func (self *Layerbeacon) Start() {

}

func (self *Layerbeacon) Stop() {

}

func (self *Layerbeacon) BeaconWatcher() {

}
