package rpcserver

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type WsServer struct {
	started          int32
	shutdown         int32
	numSocketClients int32
	config           RpcServerConfig
	httpServer       *http.Server
	statusLock       sync.RWMutex
	statusLines      map[int]string
	authSHA          []byte
	limitAuthSHA     []byte
	// channel
	cRequestProcessShutdown chan struct{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
