package wrapper

import (
	"encoding/json"
	"runtime"

	"github.com/klauspost/compress/zstd"
)

var compresser *zstd.Encoder
var decompresser *zstd.Decoder

func init() {
	compresser, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	decompresser, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(runtime.NumCPU()))
}

// EnCom: encode an interface{} to bytes and compress to shorted bytes slice
func EnCom(data interface{}) ([]byte, error) {
	//s := time.Now()
	// var buf bytes.Buffer
	// e := gob.NewEncoder(&buf)
	// err := e.Encode(data)
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	//Logger.Infof("[stream] Time Encode %v", time.Since(s).Seconds())
	//s = time.Now()
	// compresser, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	// if err != nil {
	// 	return nil, err
	// }
	res := compresser.EncodeAll(b, nil)
	//Logger.Debugf("[stream] Time Compress %v", time.Since(s).Seconds())
	//Logger.Debugf("[stream] Time %v, Len encode %v len compress %v Ratio %v", time.Since(s).Seconds(), len(b), len(res), float64(len(b))/float64(len(res)))
	return res, nil
}

// DeCom: decode bytes to an interface{}
func DeCom(data []byte, out interface{}) error {
	// decompresser, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(runtime.NumCPU()))
	// if err != nil {
	// 	return err
	// }
	rawdata, err := decompresser.DecodeAll(data, nil)
	if err != nil {
		return err
	}
	// buf := bytes.NewBuffer(rawdata)
	// d := gob.NewDecoder(buf)

	err = json.Unmarshal(rawdata, out) //d.Decode(out)
	return err
}
