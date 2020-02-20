package wrapper

import (
	"bytes"
	"encoding/gob"
	"runtime"
	"time"

	"github.com/klauspost/compress/zstd"
)

// EnCom: encode an interface{} to bytes and compress to shorted bytes slice
func EnCom(data interface{}) ([]byte, error) {
	s := time.Now()
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)
	err := e.Encode(data)
	if err != nil {
		return nil, err
	}
	b := buf.Bytes()
	compresser, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}
	res := compresser.EncodeAll(b, nil)
	Logger.Infof("[stream] Time %v, Ratio %v", time.Since(s).Seconds(), float64(len(b))/float64(len(res)))
	return res, nil
}

// DeCom: decode bytes to an interface{}
func DeCom(data []byte, out interface{}) error {
	decompresser, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(runtime.NumCPU()))
	if err != nil {
		return err
	}
	rawdata, err := decompresser.DecodeAll(data, nil)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(rawdata)
	d := gob.NewDecoder(buf)
	err = d.Decode(out)
	return err
}

// // EnCom: encode an interface{} to bytes and compress to shorted bytes slice
// func (w *Wrapper) OutEnCom(data interface{}) ([]byte, error) {
// 	jsonBytes, err := json.Marshal(data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	messageBytes, err := common.GZipFromBytes(jsonBytes)
// 	return messageBytes, nil
// }

// // DeCom: decode bytes to an interface{}
// func (w *Wrapper) OutDeCom(data []byte, out interface{}) error {
// 	jsonDecodeBytes, err := common.GZipToBytes(data)
// 	err = json.Unmarshal([]byte(jsonDecodeBytes), out)
// 	return err
// }
