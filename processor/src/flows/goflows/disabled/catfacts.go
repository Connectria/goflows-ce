// catfacts.go	- another function to generate data for testing purposes

package goflows

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GetCatFact is used to get a random cat fact to be used in testing
func GetCatFact() string {
	type catfactStuff struct {
		Fact   string `json:"fact"`
		Length int64  `json:"length"`
	}

	resp, err := http.Get("https://catfact.ninja/fact")
	if err != nil {
		return "Sadly an error occurred while attempting to retrieve the the cat fact. Error:" + err.Error()

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Sorry, but an error occurred while attempting to read response body. Error:" + err.Error()
	}
	defer resp.Body.Close()

	var fact catfactStuff
	err = json.Unmarshal(body, &fact)
	if err != nil {
		return "Sorry, but an error occurred while attempting to unmarshal JSON. Error:" + err.Error()

	}

	return fact.Fact
}
