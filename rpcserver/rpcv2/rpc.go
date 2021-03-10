package rpcv2

import (
	"context"
	"io"
)

type ServiceContainer struct{
	services serviceRegistry
}

func (s *ServiceContainer) RegisterName(name string, receiver interface{}) error {
	return s.services.registerName(name, receiver)
}

func (s ServiceContainer) HandleSingleRequest(params interface{}, method string, closeChan <-chan struct{}) (interface{}, error) {
	h := newHandler(context.Background(), &s.services)
	h.allowSubscribe = false
	defer h.close(io.EOF, nil)
	return h.handleCall(params, method)
}