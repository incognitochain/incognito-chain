package config

import (
	"encoding/json"
	"log"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/spf13/viper"
)

var ReadUnifiedToken bool

type Vault struct {
	State struct {
		Decimal uint `mapstructure:"Decimal"`
	} `mapstructure:"State"`
	TokenID string `mapstructure:"TokenID"`
}

func LoadUnifiedToken(data []byte) map[string]map[string]Vault {
	res := make(map[string]map[string]Vault)
	network := c.Network()
	//read config from file
	viper.SetConfigName(utils.GetEnv(UnifiedTokenFileKey, DefaultUnifiedTokenFile))           // name of config file (without extension)
	viper.SetConfigType(utils.GetEnv(ConfigFileTypeKey, DefaultUnifiedTokenFileType))         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		} else { //if file not found
			log.Println("Using default unified token for " + network)
			json.Unmarshal(data, &res)
		}
	} else {
		err = viper.Unmarshal(&res)
		if err != nil {
			panic(err)
		}
	}
	if _, found := res[common.PRVIDStr]; found {
		panic("Found PRV in list unified token config")
	}
	return res
}

var mainnetUnifiedToken = []byte{}
var testnet1UnifiedToken = []byte{}
var testnet2UnifiedToken = []byte{}
var localUnifiedToken = []byte{}
