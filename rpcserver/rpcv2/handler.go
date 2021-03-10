// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rpcv2

import (
	"context"
	"encoding/json"
	"reflect"
	// "strconv"
	// "strings"
	"sync"
	"errors"

	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

// handler handles JSON-RPC messages. There is one handler per connection. Note that
// handler is not safe for concurrent use. Message handling never blocks indefinitely
// because RPCs are processed on background goroutines launched by handler.
//
// The entry points for incoming messages are:
//
//    h.handleMsg(message)
//    h.handleBatch(message)
//
// Outgoing calls use the requestOp struct. Register the request before sending it
// on the connection:
//
//    op := &requestOp{ids: ...}
//    h.addRequestOp(op)
//
// Now send the request, then wait for the reply to be delivered through handleMsg:
//
//    if err := op.wait(...); err != nil {
//        h.removeRequestOp(op) // timeout, etc.
//    }
//
type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *rpcservice.RPCError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type requestOp struct {
	ids  []json.RawMessage
	err  error
	resp chan *jsonrpcMessage // receives up to len(ids) responses
}

type handler struct {
	reg            *serviceRegistry
	unsubscribeCb  *callback
	respWait       map[string]*requestOp          // active client requests
	callWG         sync.WaitGroup                 // pending call goroutines
	rootCtx        context.Context                // canceled by close()
	cancelRoot     func()                         // cancel function for rootCtx
	allowSubscribe bool
	subLock    sync.Mutex
	serverSubs map[ID]*Subscription
}

type callProc struct {
	ctx       context.Context
}

func newHandler(connCtx context.Context, reg *serviceRegistry) *handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &handler{
		reg:            reg,
		respWait:       make(map[string]*requestOp),
		rootCtx:        rootCtx,
		cancelRoot:     cancelRoot,
		allowSubscribe: false,
		serverSubs:     make(map[ID]*Subscription),
	}
	return h
}

// close cancels all requests except for inflightReq and waits for
// call goroutines to shut down.
func (h *handler) close(err error, inflightReq *requestOp) {
	h.callWG.Wait()
	h.cancelRoot()
	// h.cancelServerSubscriptions(err)
}

// handleCall processes method calls.
func (h *handler) handleCall(params interface{}, method string) (interface{}, error) {
	var callb = h.reg.callback(method)

	if callb == nil {
		return nil, errors.New("Method not found: " + method)
	}
	var rawParams json.RawMessage
	var err error
	rawParams, err = json.Marshal(params)
	if err != nil {
		return nil, errors.New("Invalid parameters")
	}
	args, err := parsePositionalArguments(rawParams, callb.argTypes)
	if err != nil {
		return nil, errors.New("Invalid parameters")
	}

	return h.runMethod(h.rootCtx, method, callb, args)
}

// runMethod runs the Go callback for an RPC method.
func (h *handler) runMethod(ctx context.Context, method string, callb *callback, args []reflect.Value) (interface{}, error) {
	return callb.call(ctx, method, args)
}