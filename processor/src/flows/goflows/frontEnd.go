// frontend.go	- function to interact with front end API

package goflows

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func frontEndAPI(method, path string, payload []byte) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(method, os.Getenv("FRONTEND_URL")+path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Key", os.Getenv("FRONTEND_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("frontEndAPI() returned status code: %v (method = '%v'; path = '%v', payload = '%v')", resp.StatusCode, method, path, payload)
	}

	return ioutil.ReadAll(resp.Body)
}
