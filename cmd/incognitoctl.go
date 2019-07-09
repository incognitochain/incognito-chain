package main

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"log"
	"strconv"
	"strings"
)

var (
	cfg *params
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "0.0.1")

	// load component
	tcfg, err := loadParams()
	if err != nil {
		log.Println("Parse component error", err.Error())
		return
	}
	cfg = tcfg

	log.Printf("Process cmd: %s", cfg.Command)
	if ok, err := common.SliceExists(CmdList, cfg.Command); ok || err == nil {
		switch cfg.Command {
		case getPrivacyTokenID:
			{
				log.Printf("Params %+v", cfg)
				if cfg.PNetwork == "" {
					log.Println("Wrong param")
					return
				}
				if cfg.PToken == "" {
					log.Println("Wrong param")
					return
				}
				tokenID := common.Hash{}

				hashPNetWork := common.HashH([]byte(cfg.PNetwork))
				log.Printf("hashPNetWork: %+v\n", hashPNetWork.String())
				copy(tokenID[:16], hashPNetWork[:16])
				log.Printf("tokenID: %+v\n", tokenID.String())

				hashPToken := common.HashH([]byte(cfg.PToken))
				log.Printf("hashPToken: %+v\n", hashPToken.String())
				copy(tokenID[16:], hashPToken[:16])

				log.Printf("Result tokenID: %+v\n", tokenID.String())
			}
		case createWalletCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
					log.Println("Wrong param")
					return
				}
				err := createWallet()
				if err != nil {
					log.Println(err)
					return
				}
			}
		case listWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
					log.Println("Wrong param")
					return
				}
				accounts, err := listAccounts()
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(accounts)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		case getWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
					log.Println("Wrong param")
					return
				}
				account, err := getAccount(cfg.WalletAccountName)
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(account)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		case createWalletAccountCmd:
			{
				if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
					log.Println("Wrong param")
					return
				}
				var shardID *byte
				if cfg.ShardID > -1 {
					temp := byte(cfg.ShardID)
					shardID = &temp
				}
				account, err := createAccount(cfg.WalletAccountName, shardID)
				if err != nil {
					log.Println(err)
					return
				}
				result, err := parseToJsonString(account)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(result))
			}
		case backupChain:
			{
				if cfg.Beacon == false && cfg.ShardIDs == "" {
					log.Println("No Expected Params")
					return
				}
				var shardIDs = []byte{}
				if cfg.ShardIDs != "" {
					strs := strings.Split(cfg.ShardIDs,",")
					if len(strs) > 256 {
						log.Println("Number of shard id to process exceed limit")
						return
					}
					for _, value := range strs {
						temp, err := strconv.Atoi(value)
						if err != nil {
							log.Println("ShardID Params MUST contain number only in range 0-255")
							return
						}
						if temp > 256 {
							log.Println("ShardID exceed MAX value (> 255)")
							return
						}
						shardID := byte(temp)
						if common.IndexOfByte(shardID, shardIDs) > 0 {
							continue
						}
						shardIDs = append(shardIDs, shardID)
					}
				//backup shard
					for _, shardID := range shardIDs {
						err := BackupShardChain(shardID, cfg.ChainDataDir, cfg.OutDataDir)
						if err != nil {
							log.Printf("Shard %+v back up failed, err %+v", shardID, err)
						}
					}
				}
				if cfg.Beacon {
					err := BackupBeaconChain(cfg.ChainDataDir, cfg.FileName)
					if err != nil {
						log.Printf("Beacon Beackup failed, err %+v", err)
					}
				}
			}
		case restoreChain:
			{
				if cfg.Beacon == false && cfg.ShardIDs == "" {
					log.Println("No Expected Params")
					return
				}
				if cfg.FileName == "" || !strings.HasSuffix(cfg.FileName, ".gz"){
					log.Println("No Expected Filename or filename format should end with .gz")
					return
				}
				var shardIDs = []byte{}
				if cfg.ShardIDs != "" {
					strs := strings.Split(cfg.ShardIDs,",")
					if len(strs) > 256 {
						log.Println("Number of shard id to process exceed limit")
						return
					}
					for _, value := range strs {
						temp, err := strconv.Atoi(value)
						if err != nil {
							log.Println("ShardID Params MUST contain number only in range 0-255")
							return
						}
						if temp > 256 {
							log.Println("ShardID exceed MAX value (> 255)")
							return
						}
						shardID := byte(temp)
						if common.IndexOfByte(shardID, shardIDs) > 0 {
							continue
						}
						shardIDs = append(shardIDs, shardID)
					}
					//backup shard
					for _, shardID := range shardIDs {
						err := RestoreShardChain(shardID, cfg.ChainDataDir, cfg.FileName)
						if err != nil {
							log.Printf("Shard %+v back up failed, err %+v", shardID, err)
						}
					}
				}
				if cfg.Beacon {
					err := RestoreBeaconChain(cfg.ChainDataDir, cfg.FileName)
					if err != nil {
						log.Printf("Beacon Restore failed, err %+v", err)
					}
				}
			}
		}
	} else {
		log.Println("Parse component error", err.Error())
	}
}

func parseToJsonString(data interface{}) ([]byte, error) {
	result, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}
