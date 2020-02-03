// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

import rpcclient "github.com/incognitochain/incognito-chain/peerv2/rpcclient"

// HighwayDiscoverer is an autogenerated mock type for the HighwayDiscoverer type
type HighwayDiscoverer struct {
	mock.Mock
}

// DiscoverHighway provides a mock function with given fields: discoverPeerAddress, shardsStr
func (_m *HighwayDiscoverer) DiscoverHighway(discoverPeerAddress string, shardsStr []string) (map[string][]rpcclient.HighwayAddr, error) {
	ret := _m.Called(discoverPeerAddress, shardsStr)

	var r0 map[string][]rpcclient.HighwayAddr
	if rf, ok := ret.Get(0).(func(string, []string) map[string][]rpcclient.HighwayAddr); ok {
		r0 = rf(discoverPeerAddress, shardsStr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]rpcclient.HighwayAddr)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []string) error); ok {
		r1 = rf(discoverPeerAddress, shardsStr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
