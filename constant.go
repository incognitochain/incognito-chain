package main

import (
	"fmt"
	"github.com/ninjadotorg/constant/database"
	_ "github.com/ninjadotorg/constant/database/lvdb"
	"github.com/ninjadotorg/constant/limits"
	"github.com/ninjadotorg/constant/wallet"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
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

	// Create db and use it.
	db, err := database.Open("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	if err != nil {
		Logger.log.Error("could not open connection to leveldb")
		Logger.log.Error(err)
		panic(err)
	}

	// Check wallet and start it
	var walletObj *wallet.Wallet
	if cfg.Wallet == true {
		walletObj = &wallet.Wallet{}
		walletObj.Config = &wallet.WalletConfig{
			DataDir:        cfg.DataDir,
			DataFile:       cfg.WalletName,
			DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
			IncrementalFee: 0, // 0 mili constant
		}
		err = walletObj.LoadWallet(cfg.WalletPassphrase)
		if err != nil {
			/*if true {*/
			// in case light mode, create wallet automatically if it not exist
			walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
			walletObj.Save(cfg.WalletPassphrase)
			/*} else {
				// write log and exit when can not load wallet
				Logger.log.Criticalf("Can not load wallet with %s. Please use constantctl to create a new wallet", walletObj.Config.DataPath)
				return err
			}*/
		}
	}

	// Create server and start it.
	server := Server{}
	server.wallet = walletObj
	err = server.NewServer(cfg.Listeners, db, activeNetParams.Params, version, interrupt)
	if err != nil {
		Logger.log.Errorf("Unable to start server on %+v", cfg.Listeners)
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

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Block and transaction processing can cause bursty allocations.  This
	// limits the garbage collector from excessively overallocating during
	// bursts.  This value was arrived at with the help of profiling live
	// usage.
	debug.SetGCPercent(10)

	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %+v\n", err)
		os.Exit(1)
	}

	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isService {
			os.Exit(0)
		}
	}

	// Work around defer not working after os.Exit()
	if err := mainMaster(nil); err != nil {
		os.Exit(1)
	}
}
