package common

import (
	"encoding/base64"
	id "github.com/dgryski/go-identicon"
	lru "github.com/hashicorp/golang-lru"
)

var icon id.Renderer
var iconCache *lru.Cache

func init() {
	key := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	icon = id.New7x7(key)
	iconCache, _ = lru.New(1000)
}

func Render(data []byte) string {
	value, exist := iconCache.Get(string(data))
	if exist {
		return value.(string)
	}
	encodeData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(icon.Render(data))
	iconCache.Add(string(data), encodeData)
	return encodeData
}
