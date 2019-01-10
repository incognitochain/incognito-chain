package common

import (
	"compress/gzip"
	"bytes"
	"io/ioutil"
)

func GZipToBytes(src []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(src); err != nil {
		return []byte{}, err
	}
	if err := gz.Flush(); err != nil {
		return []byte{}, err
	}
	if err := gz.Close(); err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}

func GZipFromBytes(src []byte) ([]byte, error) {
	var br bytes.Buffer
	br.Write(src)
	gz, err := gzip.NewReader(&br)
	if err != nil {
		return []byte{}, err
	}
	resultBytes, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return []byte{}, err
	}
	return resultBytes, nil
}
