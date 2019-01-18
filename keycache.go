package main

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
This is a utility to save data key-value into file
*/
type KeyCache struct {
	filePath string
	data     map[string]interface{}
}

func (keyCache *KeyCache) Load(filePath string) error {
	keyCache.data = map[string]interface{}{}
	keyCache.filePath = filePath

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	r, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("%s error opening file: %v", filePath, err)
	}
	defer r.Close()

	dec := json.NewDecoder(r)
	err = dec.Decode(&keyCache.data)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", filePath, err)
	}

	return nil
}

func (keyCache *KeyCache) Save() error {
	w, err := os.Create(keyCache.filePath)
	if err != nil {
		Logger.log.Infof("Error opening file %s: %v", keyCache.filePath, err)
		return err
	}

	enc := json.NewEncoder(w)
	defer w.Close()
	if err := enc.Encode(&keyCache.data); err != nil {
		Logger.log.Infof("Failed to encode file %s: %v", keyCache.filePath, err)
		return err
	}

	return nil
}

func (keyCache *KeyCache) Set(key string, value interface{}) error {
	keyCache.data[key] = value
	return nil
}

func (keyCache *KeyCache) Get(key string) interface{} {
	value, ok := keyCache.data[key]
	if ok {
		return value
	} else {
		return nil
	}
}
