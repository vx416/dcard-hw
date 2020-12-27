package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func GetResponseData(record *httptest.ResponseRecorder, target interface{}) error {
	data, err := ioutil.ReadAll(record.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func BuildRequest(method, path string, data interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var (
		req    *http.Request
		reader io.Reader
	)

	if data != nil {
		switch data.(type) {
		case []byte:
			reader = bytes.NewReader(data.([]byte))
		case map[string]interface{}:
			jsonData, err := json.Marshal(data.(map[string]interface{}))
			if err != nil {
				panic(err)
			}
			reader = bytes.NewReader(jsonData)
		default:
			jsonData, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			reader = bytes.NewReader(jsonData)
		}
		req = httptest.NewRequest(method, path, reader)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "::1234")
	return req, httptest.NewRecorder()
}
