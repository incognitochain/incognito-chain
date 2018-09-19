package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func processRequest(method string, endpoint string, mapData map[string]interface{}) (error, map[string]interface{}) {
	jsonValue, _ := json.Marshal(mapData)

	request, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonValue))

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err, nil
	}

	b, _ := ioutil.ReadAll(response.Body)

	var data map[string]interface{}
	json.Unmarshal(b, &data)
	return nil, data
}

func Get(endpoint string, data map[string]interface{}) (error, map[string]interface{}) {
	return processRequest("GET", endpoint, data)
}

func Post(endpoint string, data map[string]interface{}) (error, map[string]interface{}) {
	return processRequest("GET", endpoint, data)
}