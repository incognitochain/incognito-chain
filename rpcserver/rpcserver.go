package rpcserver

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/big0t/constant-chain/addrmanager"
	"github.com/big0t/constant-chain/blockchain"
	"github.com/big0t/constant-chain/connmanager"
	"github.com/big0t/constant-chain/database"
	"github.com/big0t/constant-chain/mempool"
	"github.com/big0t/constant-chain/wallet"
	"github.com/big0t/constant-chain/wire"
	peer2 "github.com/libp2p/go-libp2p-peer"
)

const (
	rpcAuthTimeoutSeconds = 10
	RpcServerVersion      = "1.0"
)

// timeZeroVal is simply the zero value for a time.Time and is used to avoid
// creating multiple instances.
var timeZeroVal time.Time

// UsageFlag define flags that specify additional properties about the
// circumstances under which a command can be used.
type UsageFlag uint32

// rpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	started    int32
	shutdown   int32
	numClients int32

	config     RpcServerConfig
	httpServer *http.Server

	statusLock  sync.RWMutex
	statusLines map[int]string

	authSHA      [sha256.Size]byte
	limitAuthSHA [sha256.Size]byte

	// channel
	cRequestProcessShutdown chan struct{}
}

type RpcServerConfig struct {
	Listenters      []net.Listener
	ProtocolVersion string
	ChainParams     *blockchain.Params
	BlockChain      *blockchain.BlockChain
	Database        *database.DatabaseInterface
	Wallet          *wallet.Wallet
	ConnMgr         *connmanager.ConnManager
	AddrMgr         *addrmanager.AddrManager
	NodeMode        string
	Server          interface {
		// Push TxNormal Message
		PushMessageToAll(message wire.Message) error
		PushMessageToPeer(message wire.Message, id peer2.ID) error
	}

	TxMemPool     *mempool.TxPool
	RPCMaxClients int
	RPCQuirks     bool

	// Authentication
	RPCUser      string
	RPCPass      string
	RPCLimitUser string
	RPCLimitPass string
	DisableAuth  bool

	// The fee estimator keeps track of how long transactions are left in
	// the mempool before they are mined into blocks.
	FeeEstimator map[byte]*mempool.FeeEstimator
}

func (rpcServer *RpcServer) Init(config *RpcServerConfig) {
	rpcServer.config = *config
	rpcServer.statusLines = make(map[int]string)
	if config.RPCUser != "" && config.RPCPass != "" {
		login := config.RPCUser + ":" + config.RPCPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpcServer.authSHA = sha256.Sum256([]byte(auth))
	}
	if config.RPCLimitUser != "" && config.RPCLimitPass != "" {
		login := config.RPCLimitUser + ":" + config.RPCLimitPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpcServer.limitAuthSHA = sha256.Sum256([]byte(auth))
	}
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (rpcServer RpcServer) RequestedProcessShutdown() <-chan struct{} {
	return rpcServer.cRequestProcessShutdown
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (rpcServer RpcServer) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&rpcServer.numClients)+1) > rpcServer.config.RPCMaxClients {
		Logger.log.Infof("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", rpcServer.config.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// Start is used by server.go to start the rpc listener.
func (rpcServer *RpcServer) Start() error {
	if atomic.AddInt32(&rpcServer.started, 1) != 1 {
		return NewRPCError(ErrAlreadyStarted, nil)
	}
	rpcServeMux := http.NewServeMux()
	rpcServer.httpServer = &http.Server{
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}

	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rpcServer.RpcHandleRequest(w, r)
	})
	for _, listen := range rpcServer.config.Listenters {
		go func(listen net.Listener) {
			Logger.log.Infof("RPC server listening on %s", listen.Addr())
			go rpcServer.httpServer.Serve(listen)
			Logger.log.Infof("RPC listener done for %s", listen.Addr())
		}(listen)
	}
	rpcServer.started = 1
	return nil
}

// Stop is used by server.go to stop the rpc listener.
func (rpcServer RpcServer) Stop() {
	if atomic.AddInt32(&rpcServer.shutdown, 1) != 1 {
		Logger.log.Info("RPC server is already in the process of shutting down")
	}
	Logger.log.Info("RPC server shutting down")
	if rpcServer.started != 0 {
		rpcServer.httpServer.Close()
	}
	for _, listen := range rpcServer.config.Listenters {
		listen.Close()
	}
	Logger.log.Warn("RPC server shutdown complete")
	rpcServer.started = 0
	rpcServer.shutdown = 1
}

/*
Handle all request to rpcserver
*/
func (rpcServer RpcServer) RpcHandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	r.Close = true

	// Limit the number of connections to max allowed.
	if rpcServer.limitConnections(w, r.RemoteAddr) {
		return
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Keep track of the number of connected clients.
	rpcServer.IncrementClients()
	defer rpcServer.DecrementClients()
	// Check authentication for rpc user
	ok, isLimitUser, err := rpcServer.checkAuth(r, true)
	if err != nil || !ok {
		Logger.log.Error(err)
		rpcServer.AuthFail(w)
		return
	}

	rpcServer.ProcessRpcRequest(w, r, isLimitUser)
}

// checkAuth checks the HTTP Basic authentication supplied by a wallet
// or RPC client in the HTTP request r.  If the supplied authentication
// does not match the username and password expected, a non-nil error is
// returned.
//
// This check is time-constant.
//
// The first bool return value signifies auth success (true if successful) and
// the second bool return value specifies whether the user can change the state
// of the server (true) or whether the user is limited (false). The second is
// always false if the first is.
func (rpcServer RpcServer) checkAuth(r *http.Request, require bool) (bool, bool, error) {
	if rpcServer.config.DisableAuth {
		return true, true, nil
	}
	authhdr := r.Header["Authorization"]
	if len(authhdr) <= 0 {
		if require {
			Logger.log.Warnf("RPC authentication failure from %s",
				r.RemoteAddr)
			return false, false, errors.New("auth failure")
		}

		return false, false, nil
	}

	authsha := sha256.Sum256([]byte(authhdr[0]))

	// Check for limited auth first as in environments with limited users, those
	// are probably expected to have a higher volume of calls
	limitcmp := subtle.ConstantTimeCompare(authsha[:], rpcServer.limitAuthSHA[:])
	if limitcmp == 1 {
		return true, true, nil
	}

	// Check for admin-level auth
	cmp := subtle.ConstantTimeCompare(authsha[:], rpcServer.authSHA[:])
	if cmp == 1 {
		return true, false, nil
	}

	// RpcRequest's auth doesn't match either user
	Logger.log.Warnf("RPC authentication failure from %s", r.RemoteAddr)
	return false, false, NewRPCError(ErrAuthFail, nil)
}

// IncrementClients adds one to the number of connected RPC clients.  Note
// this only applies to standard clients.
//
// This function is safe for concurrent access.
func (rpcServer *RpcServer) IncrementClients() {
	atomic.AddInt32(&rpcServer.numClients, 1)
}

// DecrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.
//
// This function is safe for concurrent access.
func (rpcServer *RpcServer) DecrementClients() {
	atomic.AddInt32(&rpcServer.numClients, -1)
}

// AuthFail sends a Message back to the client if the http auth is rejected.
func (rpcServer RpcServer) AuthFail(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", `Basic realm="RPC"`)
	http.Error(w, "401 Unauthorized.", http.StatusUnauthorized)
}

/*
handles reading and responding to RPC messages.
*/
func (rpcServer RpcServer) ProcessRpcRequest(w http.ResponseWriter, r *http.Request, isLimitedUser bool) {
	if atomic.LoadInt32(&rpcServer.shutdown) != 0 {
		return
	}
	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON Message: %+v", errCode, err), errCode)
		return
	}
	// Logger.log.Info(string(body))
	log.Println(string(body))

	// Unfortunately, the http server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		Logger.log.Error(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		Logger.log.Errorf("Failed to hijack HTTP connection: %s", err.Error())
		Logger.log.Error(err)
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
	var request RpcRequest
	if err := json.Unmarshal(body, &request); err != nil {
		jsonErr = NewRPCError(ErrRPCParse, err)
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
		// Rpc does not respond to any request without and "id" or "id":null,
		// regardless the indicated JSON-RPC protocol version unless RPC quirks
		// are enabled. With RPC quirks enabled, such requests will be responded
		// to if the reqeust does not indicate JSON-RPC version.
		//
		// RPC quirks can be enabled by the user to avoid compatibility issues
		// with software relying on Core's behavior.
		if request.Id == nil && !(rpcServer.config.RPCQuirks && request.Jsonrpc == "") {
			return
		}

		// The parse was at least successful enough to have an Id so
		// set it for the response.
		responseID = request.Id

		// Setup a close notifier.  Since the connection is hijacked,
		// the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()

		// Check if the user is limited and set error if method unauthorized
		if !isLimitedUser {
			if function, ok := RpcLimited[request.Method]; ok {
				_ = function
				jsonErr = NewRPCError(ErrRPCInvalidMethodPermission, errors.New(""))
			}
		}
		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete
			// command.
			command := RpcHandler[request.Method]
			if command == nil {
				if isLimitedUser {
					command = RpcLimited[request.Method]
				} else {
					result = nil
					jsonErr = NewRPCError(ErrRPCMethodNotFound, nil)
				}
			}
			if command != nil {
				result, jsonErr = command(rpcServer, request.Params, closeChan)
			} else {
				jsonErr = NewRPCError(ErrRPCMethodNotFound, nil)
			}
		}
	}
	if jsonErr.(*RPCError) != nil && r.Method != "OPTIONS" {
		// Logger.log.Errorf("RPC function process with err \n %+v", jsonErr)
		log.Printf("RPC function process with err \n %+v", jsonErr)
	}
	// Marshal the response.
	msg, err := rpcServer.createMarshalledReply(responseID, result, jsonErr)
	if err != nil {
		Logger.log.Errorf("Failed to marshal reply: %s", err.Error())
		Logger.log.Error(err)
		return
	}

	// Write the response.
	err = rpcServer.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		Logger.log.Errorf("Failed to write marshalled reply: %s", err.Error())
		Logger.log.Error(err)
	}

	// Terminate with newline to maintain compatibility with coin Core.
	if err := buf.WriteByte('\n'); err != nil {
		Logger.log.Errorf("Failed to append terminating newline to reply: %s", err.Error())
		Logger.log.Error(err)
	}
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert errors that are not of
// the type *btcjson.RPCError to the appropriate type as needed.
func (rpcServer RpcServer) createMarshalledReply(id, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = rpcServer.internalRPCError(replyErr.Error(), "")
		}
	}

	return MarshalResponse(id, result, jsonErr)
}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate Code set.  It also logs the error to the
// RPC server subsystem since internal errors really should not occur.  The
// context parameter is only used in the log Message and may be empty if it's
// not needed.
func (rpcServer RpcServer) internalRPCError(errStr, context string) *RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	Logger.log.Info(logStr)
	return NewRPCError(ErrRPCInternal, errors.New(errStr))
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status Code.  This function was lifted and
// adapted from the standard library HTTP server Code since it's not exported.
func (rpcServer RpcServer) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	rpcServer.statusLock.RLock()
	line, ok := rpcServer.statusLines[key]
	rpcServer.statusLock.RUnlock()
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
		rpcServer.statusLock.Lock()
		rpcServer.statusLines[key] = line
		rpcServer.statusLock.Unlock()
	} else {
		text = "status Code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}

	return line
}

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, a status Code, and a writer.
func (rpcServer RpcServer) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, rpcServer.httpStatusLine(req, code))
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
