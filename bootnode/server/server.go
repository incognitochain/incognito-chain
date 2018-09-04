package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ninjadotorg/cash-prototype/bootnode/server/jsonrpc"
	"github.com/ninjadotorg/cash-prototype/common"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	rpcAuthTimeoutSeconds = 10
	heartbeatTimeout = 5
)

// timeZeroVal is simply the zero value for a time.Time and is used to avoid
// creating multiple instances.
var timeZeroVal time.Time

// UsageFlag define flags that specify additional properties about the
// circumstances under which a command can be used.
type UsageFlag uint32

type Peer struct {
	ID string
	FirstPing time.Time
	LastPing time.Time
}

// rpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	started    int32
	shutdown   int32
	numClients int32

	peers []Peer

	Config     RpcServerConfig
	HttpServer *http.Server

	statusLock  sync.RWMutex
	statusLines map[int]string

	requestProcessShutdown chan struct{}
	quit                   chan int
}

type RpcServerConfig struct {
	Listeners []net.Listener
	RPCMaxClients int
}

func (self *RpcServer) Init(config *RpcServerConfig) (error) {
	self.Config = *config
	self.statusLines = make(map[int]string)
	self.peers = make([]Peer, 0)
	go self.PeerHeartBeat()
	return nil
}

func (self *RpcServer) AddPeer(ID string) {
	exist := false
	for _, peer := range self.peers {
		if peer.ID == ID {
			exist = true
			peer.LastPing = time.Now().Local()
		}
	}

	if !exist {
		self.peers = append(self.peers, Peer{ID, time.Now().Local(), time.Now().Local()})
	}
}

func (self *RpcServer) RemovePeer(ID string) {
	removeIdx := -1
	for idx, peer := range self.peers {
		if peer.ID == ID {
			removeIdx = idx
		}
	}

	if removeIdx != -1 {
		self.RemovePeerByIdx(removeIdx)
	}
}

func (self *RpcServer) RemovePeerByIdx(idx int) {
	self.peers = append(self.peers[:idx], self.peers[idx+1:]...)
}

func (self *RpcServer) PeerHeartBeat() {
	for {
		now := time.Now().Local()
		if len(self.peers) > 0 {
		loop:
			for idx, peer := range self.peers {
				if now.Sub(peer.LastPing).Seconds() > heartbeatTimeout {
					self.RemovePeerByIdx(idx)
					goto loop
				}
			}
		}
		time.Sleep(heartbeatTimeout * time.Second)
	}
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (self RpcServer) RequestedProcessShutdown() <-chan struct{} {
	return self.requestProcessShutdown
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (self RpcServer) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&self.numClients)+1) > self.Config.RPCMaxClients {
		Logger.log.Infof("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", self.Config.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

func (self RpcServer) Start() error {
	if atomic.AddInt32(&self.started, 1) != 1 {
		return errors.New("RPC server is already started")
	}
	rpcServeMux := http.NewServeMux()
	self.HttpServer = &http.Server{
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}

	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		self.RpcHandleRequest(w, r)
	})

	for _, listen := range self.Config.Listeners {
		go func(listen net.Listener) {
			Logger.log.Infof("RPC server listening on %s", listen.Addr())
			go self.HttpServer.Serve(listen)
			Logger.log.Infof("RPC listener done for %s", listen.Addr())
		}(listen)
	}

	self.started = 1
	return nil
}

// Stop is used by server.go to stop the rpc listener.
func (self RpcServer) Stop() error {
	if atomic.AddInt32(&self.shutdown, 1) != 1 {
		Logger.log.Info("RPC server is already in the process of shutting down")
		return nil
	}
	Logger.log.Info("RPC server shutting down")
	if self.started != 0 {
		self.HttpServer.Close()
		close(self.quit)
	}
	for _, listen := range self.Config.Listeners {
		listen.Close()
	}
	Logger.log.Info("RPC server shutdown complete")
	self.started = 0
	self.shutdown = 1
	return nil
}

func (self RpcServer) RpcHandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
	r.Close = true

	// Limit the number of connections to max allowed.
	if self.limitConnections(w, r.RemoteAddr) {
		return
	}

	// Keep track of the number of connected clients.
	self.IncrementClients()
	defer self.DecrementClients()
	self.ProcessRpcRequest(w, r)
}

// IncrementClients adds one to the number of connected RPC clients.  Note
// this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (self *RpcServer) IncrementClients() {
	atomic.AddInt32(&self.numClients, 1)
}

// DecrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (self *RpcServer) DecrementClients() {
	atomic.AddInt32(&self.numClients, -1)
}
/**
handles reading and responding to RPC messages.
*/
func (self RpcServer) ProcessRpcRequest(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&self.shutdown) != 0 {
		return
	}
	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v",
			errCode, err), errCode)
		return
	}
	Logger.log.Info(string(body))

	// Unfortunately, the http server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.

	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		log.Print(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		Logger.log.Infof("Failed to hijack HTTP connection: %v", err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()
	conn.SetReadDeadline(timeZeroVal)

	// Attempt to parse the raw body into a JSON-RPC request.
	var responseID interface{}
	var jsonErr error
	var result interface{}
	var request jsonrpc.Request
	if err := json.Unmarshal(body, &request); err != nil {
		jsonErr = &common.RPCError{
			Code:    common.ErrRPCParse.Code,
			Message: "Failed to parse request: " + err.Error(),
		}
	}

	if request.Method == "" {
		// Write the response.
		msg, err := self.createMarshalledReply(responseID, result, jsonErr)
		if err != nil {
			Logger.log.Infof("Failed to marshal reply: %v", err)
			return
		}
		err = self.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
		if err != nil {
			Logger.log.Info(err)
			return
		}
		if _, err := buf.Write(msg); err != nil {
			Logger.log.Infof("Failed to write marshalled reply: %v", err)
		}
		// Terminate with newline to maintain compatibility with coin Core.
		if err := buf.WriteByte('\n'); err != nil {
			Logger.log.Infof("Failed to append terminating newline to reply: %v", err)
		}
		return
	}

	if jsonErr == nil {
		// The JSON-RPC 1.0 spec defines that notifications must have their "id"
		// set to null and states that notifications do not have a response.
		//
		// A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and
		// without an "id" member. The specification states that notifications
		// must not be responded to. JSON-RPC 2.0 permits the null value as a
		// valid request id, therefore such requests are not notifications.
		//
		// coin Core serves requests with "id":null or even an absent "id",
		// and responds to such requests with "id":null in the response.
		//
		// Btcd does not respond to any request without and "id" or "id":null,
		// regardless the indicated JSON-RPC protocol version unless RPC quirks
		// are enabled. With RPC quirks enabled, such requests will be responded
		// to if the reqeust does not indicate JSON-RPC version.
		//
		// RPC quirks can be enabled by the user to avoid compatibility issues
		// with software relying on Core's behavior.
		if request.ID == nil {
			return
		}

		// The parse was at least successful enough to have an ID so
		// set it for the response.
		responseID = request.ID

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()

		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			command := RpcHandler[request.Method]
			result, jsonErr = command(self, request.Params, closeChan)
		}
	}
	// Marshal the response.
	msg, err := self.createMarshalledReply(responseID, result, jsonErr)
	if err != nil {
		Logger.log.Infof("Failed to marshal reply: %v", err)
		return
	}

	// Write the response.
	err = self.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		Logger.log.Info(err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		Logger.log.Infof("Failed to write marshalled reply: %v", err)
	}

	// Terminate with newline to maintain compatibility with coin Core.
	if err := buf.WriteByte('\n'); err != nil {
		Logger.log.Infof("Failed to append terminating newline to reply: %v", err)
	}
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func (self RpcServer) createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *common.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*common.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = self.internalRPCError(replyErr.Error(), "")
		}
	}

	return jsonrpc.MarshalResponse(id, result, jsonErr)
}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log message and may be empty if it's
// not needed.
func (self RpcServer) internalRPCError(errStr, context string) *common.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	Logger.log.Info(logStr)
	return common.NewRPCError(common.ErrRPCInternal.Code, errStr)
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status code.  This function was lifted and
// adapted from the standard library HTTP server code since it's not exported.
func (self RpcServer) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	self.statusLock.RLock()
	line, ok := self.statusLines[key]
	self.statusLock.RUnlock()
	if ok {
		return line
	}

	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		self.statusLock.Lock()
		self.statusLines[key] = line
		self.statusLock.Unlock()
	} else {
		text = "status code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}

	return line
}

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, a status code, and a writer.
func (self RpcServer) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, self.httpStatusLine(req, code))
	if err != nil {
		return err
	}

	err = headers.Write(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}
