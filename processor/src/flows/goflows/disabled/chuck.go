// chuck.go	- functions to generate test data

package goflows

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GetChuck used to get a random joke for testing purposes
func GetChuck() string {
	type chuckStuff struct {
		Value string `json:"value"`
	}

	resp, err := http.Get("https://api.chucknorris.io/jokes/random")
	if err != nil {
		return "Sorry, but an error occurred while attempting to retrieve the joke. Error:" + err.Error()

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Sorry, but an error occurred while attempting to read response body. Error:" + err.Error()
	}
	defer resp.Body.Close()

	var joke chuckStuff
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "Sorry, but an error occurred while attempting to unmarshal JSON. Error:" + err.Error()

	}

	return joke.Value
}
