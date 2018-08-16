package rpcserver

import (
	"github.com/internet-cash/prototype/blockchain"
	"sync/atomic"
	"net/http"
	"errors"
	"time"
	"log"
	"io/ioutil"
	"os"
)

const (
	rpcAuthTimeoutSeconds = 10
)

type commandHandler func(*RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"createtransaction": handleCreateTransaction,
}

// rpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	started    int32
	numClients int32

	Config RpcServerConfig
}

type RpcServerConfig struct {
	ChainParams   *blockchain.Params
	RPCMaxClients int
	Port          string
}

func (self RpcServer) Init(config *RpcServerConfig) (*RpcServer, error) {
	self.Config = *config
	return &self, nil
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (self RpcServer) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&self.numClients)+1) > self.Config.RPCMaxClients {
		log.Printf("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", self.Config.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// genCertPair generates a key/cert pair to the paths provided.
func genCertPair(certFile, keyFile string) error {
	log.Println("Generating TLS certificates...")

	org := "btcd autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := btcutil.NewTLSCertPair(org, validUntil, nil)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	rpcsLog.Infof("Done generating TLS certificates")
	return nil
}

func (self RpcServer) Start() (error) {
	if atomic.AddInt32(&self.started, 1) != 1 {
		return errors.New("RPC server is already started")
	}
	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    ":" + self.Config.Port,
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}

	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true
	})
	httpServer.ListenAndServe()
	return nil
}

func handleCreateTransaction(self *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, nil
}
