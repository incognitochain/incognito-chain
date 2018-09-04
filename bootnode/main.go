package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"github.com/ninjadotorg/cash-prototype/bootnode/server"
	"github.com/ninjadotorg/cash-prototype/common"
	"runtime"
	"strings"
)

var (
	cfg *config
)

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string, netType string) ([]common.SimpleAddr, error) {
	netAddrs := make([]common.SimpleAddr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			//netAddrs = append(netAddrs, simpleAddr{net: netType + "6", addr: addr})
		} else {
			netAddrs = append(netAddrs, common.SimpleAddr{Net: netType + "4", Addr: addr})
		}
	}
	return netAddrs, nil
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen

	netAddrs, err := parseListeners(cfg.RPCListeners, "tcp")
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			log.Printf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

func Start(interrupt <-chan struct{}) error {
	// load config
	tcfg, _, err := loadConfig()

	if err != nil {
		return err
	}
	cfg = tcfg

	rpcListeners, err := setupRPCListeners()
	if err != nil {
		return err
	}
	if len(rpcListeners) == 0 {
		return errors.New("RPCS: No valid listen address")
	}

	rpcConfig := server.RpcServerConfig{
		Listeners:    rpcListeners,
		RPCMaxClients: cfg.RPCMaxClients,
	}
	server := &server.RpcServer{}
	err = server.Init(&rpcConfig)
	if err != nil {
		return err
	}

	err = server.Start()
	if err != nil {
		return err
	}

	// Signal process shutdown when the RPC server requests it.
	go func() {
		<-server.RequestedProcessShutdown()
		shutdownRequestChannel <- struct{}{}
	}()



	<-interrupt
	return nil
}

func main() {
	interrupt := interruptListener()
	defer log.Println("Shutdown complete")

	// Show version at startup.
	log.Printf("Version %s", "1")

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return
	}

	// Work around defer not working after os.Exit()
	if err := Start(interrupt); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
