package rpcserver

import (
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type WsServer struct {
	started      int32
	shutdown     int32
	numWsClients int32
	config       RpcServerConfig
	server       *http.Server
	statusLock   sync.RWMutex
	authSHA      []byte
	limitAuthSHA []byte
	// channel
	cRequestProcessShutdown chan struct{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (wsServer *WsServer) Init(config *RpcServerConfig) {
	wsServer.config = *config
}

// Start is used by rpcserver.go to start the rpc listener.
func (wsServer *WsServer) Start() error {
	if atomic.AddInt32(&wsServer.started, 1) != 1 {
		return NewRPCError(ErrAlreadyStarted, nil)
	}
	wsServeMux := http.NewServeMux()
	wsServer.server = &http.Server{
		Handler: wsServeMux,
		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}
	wsServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wsServer.handleWsRequest(w, r)
	})
	for _, listen := range wsServer.config.WsListenters {
		go func(listen net.Listener) {
			Logger.log.Infof("RPC Websocket server listening on %s", listen.Addr())
			go wsServer.server.Serve(listen)
			Logger.log.Infof("RPC Websocket listener done for %s", listen.Addr())
		}(listen)
	}
	wsServer.started = 1
	return nil
}

// Stop is used by rpcserver.go to stop the rpc listener.
func (wsServer *WsServer) Stop() {
	if atomic.AddInt32(&wsServer.shutdown, 1) != 1 {
		Logger.log.Info("RPC server is already in the process of shutting down")
	}
	Logger.log.Info("RPC server shutting down")
	if wsServer.started != 0 {
		wsServer.server.Close()
	}
	for _, listen := range wsServer.config.HttpListenters {
		listen.Close()
	}
	Logger.log.Warn("RPC server shutdown complete")
	wsServer.started = 0
	wsServer.shutdown = 1
}

/*
Handle all ws request to rpcserver
*/
// @NOTICE: no auth for this version yet
func (wsServer *WsServer) handleWsRequest(w http.ResponseWriter, r *http.Request) {
	if wsServer.limitWsConnections(w, r.RemoteAddr) {
		return
	}
	// Keep track of the number of connected clients.
	wsServer.IncrementWsClients()
	defer wsServer.DecrementWsClients()

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	wsServer.ProcessRpcWsRequest(ws)
}

func (wsServer *WsServer) ProcessRpcWsRequest(ws *websocket.Conn) {
	if atomic.LoadInt32(&wsServer.shutdown) != 0 {
		return
	}
	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}
		var jsonErr error
		var cResult chan interface{}
		var cError chan *RPCError
		var result interface{}
		var subcriptionRequest *SubcriptionRequest
		subcriptionRequest, jsonErr = parseSubcriptionRequest(msg)
		if jsonErr == nil {
			request := subcriptionRequest.JsonRequest
			if request.Id == nil && !(wsServer.config.RPCQuirks && request.Jsonrpc == "") {
				return
			}
			// The parse was at least successful enough to have an Id so
			// set it for the response.
			// Setup a close notifier.  Since the connection is hijacked,
			// the CloseNotifer on the ResponseWriter is not available.
			closeChan := make(chan struct{}, 1)
			if jsonErr == nil {
				// Attempt to parse the JSON-RPC request into a known concrete
				// command.
				command := WsHandler[request.Method]
				if command == nil {
					result = nil
					jsonErr = NewRPCError(ErrRPCMethodNotFound, nil)
				} else {
					cResult = make(chan interface{})
					cError = make(chan *RPCError)
					go command(wsServer, request.Params, cResult, cError, closeChan)
					// Marshal the response.
					for result = range cResult {
						res, err := createMarshalledResponse(&request, result, jsonErr)
						if err != nil {
							Logger.log.Errorf("Failed to marshal reply: %s", err.Error())
							return
						}
						if err := ws.WriteMessage(msgType, res); err != nil {
							Logger.log.Errorf("Failed to marshal reply: %+v", err)
							return
						}
					}
				}
			}
		}
	}
}

func (wsServer *WsServer) IncrementWsClients() {
	atomic.AddInt32(&wsServer.numWsClients, 1)
}
func (wsServer *WsServer) DecrementWsClients() {
	atomic.AddInt32(&wsServer.numWsClients, -1)
}
func (wsServer *WsServer) limitWsConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&wsServer.numWsClients)+1) > wsServer.config.RPCMaxWSClients {
		Logger.log.Infof("Max RPC Web Socket exceeded [%d] - "+
			"disconnecting client %s", wsServer.config.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}
