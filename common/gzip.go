package common

import (
	"compress/gzip"
	"bytes"
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
	tmpBytes := make([]byte, 101024)
	resultBytes := make([]byte, 0)
	for {
		l, err := gz.Read(tmpBytes)
		if err != nil {
			return []byte{}, err
		}
		resultBytes = append(resultBytes, tmpBytes[:l]...)
		if l < 101024 {
			break
		}
	}
	if err := gz.Close(); err != nil {
		return []byte{}, err
	}
	return resultBytes, nil
}
