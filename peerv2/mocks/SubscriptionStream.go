// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"

import pubsub "github.com/libp2p/go-libp2p-pubsub"

// SubscriptionStream is an autogenerated mock type for the SubscriptionStream type
type SubscriptionStream struct {
	mock.Mock
}

// Next provides a mock function with given fields: _a0
func (_m *SubscriptionStream) Next(_a0 context.Context) (*pubsub.Message, error) {
	ret := _m.Called(_a0)

	var r0 *pubsub.Message
	if rf, ok := ret.Get(0).(func(context.Context) *pubsub.Message); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pubsub.Message)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
