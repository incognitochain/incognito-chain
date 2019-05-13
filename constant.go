package main

import (
	"fmt"
	"github.com/constant-money/constant-chain/common"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/constant-money/constant-chain/database"
	_ "github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/databasemp"
	_ "github.com/constant-money/constant-chain/databasemp/lvdb"
	"github.com/constant-money/constant-chain/limits"
	"github.com/constant-money/constant-chain/wallet"
)

var (
	cfg *config
)

// winServiceMain is only invoked on Windows.  It detects when constant network is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

// mainMaster is the real main function for constant network.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func mainMaster(serverChan chan<- *Server) error {
	tempConfig, _, err := loadConfig()
	if err != nil {
		log.Println("Load config error")
		log.Println(err)
		return err
	}
	cfg = tempConfig
	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer Logger.log.Warn("Shutdown complete")

	// Show version at startup.
	version := version()
	Logger.log.Infof("Version %s", version)

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	db, err := database.Open("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	// Create db and use it.
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}
	// Create db mempool and use it
	dbmp, err := databasemp.Open("leveldbmempool", filepath.Join(cfg.DataDir, cfg.DatabaseMempoolDir))
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}

	// Check wallet and start it
	var walletObj *wallet.Wallet
	if cfg.Wallet {
		walletObj = &wallet.Wallet{}
		walletObj.SetConfig(&wallet.WalletConfig{
			DataDir:        cfg.DataDir,
			DataFile:       cfg.WalletName,
			DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
			IncrementalFee: 0, // 0 mili constant
		},
		)
		err = walletObj.LoadWallet(cfg.WalletPassphrase)
		if err != nil {
			if cfg.WalletAutoInit {
				Logger.log.Critical("\n **** Auto init wallet flag is TRUE ****\n")
				walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
				walletObj.Save(cfg.WalletPassphrase)
			} else {
				// write log and exit when can not load wallet
				Logger.log.Criticalf("Can not load wallet with %s. Please use constantctl to create a new wallet", walletObj.GetConfig().DataPath)
				return err
			}
		}
	}

	// Create server and start it.
	server := Server{}
	server.wallet = walletObj
	err = server.NewServer(cfg.Listener, db, dbmp, activeNetParams.Params, version, interrupt)
	if err != nil {
		Logger.log.Errorf("Unable to start server on %+v", cfg.Listener)
		Logger.log.Error(err)
		return err
	}
	defer func() {
		Logger.log.Warn("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
		Logger.log.Warn("Server shutdown complete")
	}()
	server.Start()
	if serverChan != nil {
		serverChan <- &server
	}

	// Check Metric analyzation system
	env := os.Getenv("GrafanaURL")
	if env != "" {
		Logger.log.Criticalf("Metric Server: %+v", os.Getenv("GrafanaURL"))
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("NumCPU", runtime.NumCPU())
	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(30)
	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %+v\n", err)
		os.Exit(common.ExitByOs)
	}

	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(common.ExitByOs)
		}
		if isService {
			os.Exit(common.ExitCodeUnknow)
		}
	}

	// Work around defer not working after os.Exit()
	if err := mainMaster(nil); err != nil {
		os.Exit(common.ExitByOs)
	}
}
