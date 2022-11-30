package multiview

type BeaconMultiView struct {
	*multiView
}

func NewBeaconMultiView() *BeaconMultiView {
	bv := &BeaconMultiView{}
	bv.multiView = NewMultiView()
	bv.RunCleanProcess()
	return bv
}

func (s *BeaconMultiView) SimulateAddView(view View) (cloneMultiview MultiView) {
	sv := &BeaconMultiView{}
	sv.multiView = s.Clone().(*multiView)
	sv.AddView(view)
	return sv
}

func (s *BeaconMultiView) AddView(v View) (int, error) {
	added := s.multiView.addView(v)
	err := s.FinalizeView(*s.GetExpectedFinalView().GetHash())
	res := 0
	if added {
		res = 1
	}
	return res, err
}

func (s *BeaconMultiView) ReplaceView(v View) bool {
	return s.multiView.replaceView(v)
}
