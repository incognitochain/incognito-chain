package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func processRequest(method string, endpoint string, mapData map[string]interface{}) (error, []byte) {
	jsonValue, _ := json.Marshal(mapData)
	request, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonValue))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err, nil
	}

	b, error := ioutil.ReadAll(response.Body)

	return error, b
}

func Get(endpoint string, data map[string]interface{}) (error, []byte) {
	return processRequest("GET", endpoint, data)
}

func Post(endpoint string, data map[string]interface{}) (error, []byte) {
	return processRequest("GET", endpoint, data)
}