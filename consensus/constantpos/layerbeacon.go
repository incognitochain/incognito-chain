package constantpos

type Layerbeacon struct {
	cQuit     chan struct{}
	Committee CommitteeStruct
	started   bool
}

func (self *Layerbeacon) Start() {

}
func (self *Layerbeacon) Stop() {

}
