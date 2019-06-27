package wallet

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var dataDir string
var wallet *Wallet
var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	dataDir, _ = os.Getwd()
	wallet = new(Wallet)
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()


func TestInit(t *testing.T){
	data := []struct {
		passPhrase string
		numOfAccount uint32
		name string
	}{
		{"", uint32(2), "Wallet1"},
		{"12345678", uint32(0), "Wallet1"},
		{"12345678", uint32(3), ""},
	}

	wallet := new(Wallet)

	for _, item := range data {
		wallet.Init(item.passPhrase, item.numOfAccount, item.name)

		if item.numOfAccount == 0{
			assert.Equal(t, 1, len(wallet.MasterAccount.Child))
		} else {
			assert.Equal(t, int(item.numOfAccount), len(wallet.MasterAccount.Child))
		}

		if item.name == "" {
			assert.Equal(t, WalletNameDefault, wallet.Name)
		} else {
			assert.Equal(t, item.name, wallet.Name)
		}

		Logger.log.Infof("Wallet: %v\n", wallet)

		assert.Equal(t, item.passPhrase, wallet.PassPhrase)
		assert.Equal(t, SeedKeyLen, len(wallet.Seed))
		assert.Greater(t, len(wallet.Mnemonic), 0)
	}
}

func TestWallet_ExportAccount(t *testing.T) {

}

func TestCreateNewAccount(t *testing.T){
	data := []struct {
		accountName string
		shardID byte
	}{
		{"", byte(0)},
		{"Acc A", byte(1)},
		//{"Acc A", },
	}

	wallet := new(Wallet)
	wallet.Init("", 0, "")

	//tempConfig, _, err := loadConfig()
	//if err != nil {
	//	log.Println("Load config error")
	//	log.Println(err)
	//	return err
	//}
	//cfg = tempConfig
	//// Get a channel that will be closed when a shutdown signal has been
	//// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	//// another subsystem such as the RPC server.
	//interrupt := interruptListener()
	//defer Logger.log.Warn("Shutdown complete")
	//
	//// Show version at startup.
	//version := version()
	//Logger.log.Infof("Version %s", version)
	//
	//// Return now if an interrupt signal was triggered.
	//if interruptRequested(interrupt) {
	//	return nil
	//}
	//
	//db, err := database.Open("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	//// Create db and use it.
	//if err != nil {
	//	Logger.log.Error("could not open connection to leveldb")
	//	Logger.log.Error(err)
	//	panic(err)
	//}
	//// Create db mempool and use it
	//dbmp, err := databasemp.Open("leveldbmempool", filepath.Join(cfg.DataDir, cfg.DatabaseMempoolDir))
	//if err != nil {
	//	Logger.log.Error("could not open connection to leveldb")
	//	Logger.log.Error(err)
	//	panic(err)
	//}
	//
	//// Check wallet and start it
	//var walletObj *wallet.Wallet
	//if cfg.Wallet {
	//	walletObj = &wallet.Wallet{}
	//	walletConf := wallet.WalletConfig{
	//		DataDir:        cfg.DataDir,
	//		DataFile:       cfg.WalletName,
	//		DataPath:       filepath.Join(cfg.DataDir, cfg.WalletName),
	//		IncrementalFee: 0, // 0 mili PRV
	//	}
	//	if cfg.WalletShardID >= 0 {
	//		// check shardID of wallet
	//		temp := byte(cfg.WalletShardID)
	//		walletConf.ShardID = &temp
	//	}
	//	walletObj.SetConfig(&walletConf)
	//	err = walletObj.LoadWallet(cfg.WalletPassphrase)
	//	if err != nil {
	//		if cfg.WalletAutoInit {
	//			Logger.log.Critical("\n **** Auto init wallet flag is TRUE ****\n")
	//			walletObj.Init(cfg.WalletPassphrase, 0, cfg.WalletName)
	//			walletObj.Save(cfg.WalletPassphrase)
	//		} else {
	//			// write log and exit when can not load wallet
	//			Logger.log.Criticalf("Can not load wallet with %s. Please use incognitoctl to create a new wallet", walletObj.GetConfig().DataPath)
	//			return err
	//		}
	//	}
	//}

	numAccount := len(wallet.MasterAccount.Child)
	Logger.log.Errorf("numAccount: %v\n", numAccount)
	//fmt.Printf("numAccount: %v\n", numAccount)
	for _, item := range data {
		//fmt.Printf("item.accountName: %v\n", item.accountName)
		Logger.log.Infof("item.accountName: %v\n", item.accountName)
		wallet.CreateNewAccount(item.accountName, &item.shardID)
		newAccount := wallet.MasterAccount.Child[numAccount]

		assert.Equal(t, numAccount + 1, len(wallet.MasterAccount.Child))

		if item.accountName == "" {
			assert.Equal(t, "AccountWallet "+string(numAccount), newAccount.Name)
		}
	}


	//wallet := new(Wallet)
	//wallet.Init("123", 1, "Wallet A")
	//var shardID *byte
	//shard0 := byte(0)
	//shardID = &shard0
	//wallet.CreateNewAccount("Hien",shardID)

}

