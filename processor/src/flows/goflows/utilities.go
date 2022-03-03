// utilities.go - utility functions

package goflows

import (
	"crypto/rand"
	"fmt"
)

// generate a UUID for the "run"
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "1-2-3-4-5"
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
