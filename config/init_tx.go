package config

import (
	"encoding/json"
	"log"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/utils"
	"github.com/spf13/viper"
)

type InitialIncognito struct {
	Version              int                    `mapstructure:"Version"`
	Type                 string                 `mapstructure:"Type"`
	LockTime             uint64                 `mapstructure:"LockTime"`
	Fee                  int                    `mapstructure:"Fee"`
	Info                 string                 `mapstructure:"Info"`
	SigPubKey            string                 `mapstructure:"SigPubKey"`
	Sig                  string                 `mapstructure:"Sig"`
	Proof                string                 `mapstructure:"Proof"`
	PubKeyLastByteSender int                    `mapstructure:"PubKeyLastByteSender"`
	Metadata             map[string]interface{} `mapstructure:"Metadata"`
}

type initTx struct {
	InitialIncognito []InitialIncognito `mapstructure:"initial_incognito" description:"fee per tx calculate by kb"`
}

func (initTx *initTx) load(data []byte) {
	network := c.Network()
	//read config from file
	viper.SetConfigName(utils.GetEnv(InitTxFileKey, DefaultInitTxFile))                       // name of config file (without extension)
	viper.SetConfigType(utils.GetEnv(ConfigFileTypeKey, DefaultInitTxFileType))               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		} else { //if file not found
			log.Println("Using default init tx for " + network)
			initTxs := []InitialIncognito{}
			json.Unmarshal(data, &initTxs)
			initTx.InitialIncognito = initTxs
		}
	} else {
		err = viper.Unmarshal(&initTx)
		if err != nil {
			panic(err)
		}
	}
}
