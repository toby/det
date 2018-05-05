package server

import (
	"encoding/json"
)

const DetApiVersion = "0.1"

type DetSemaphore struct {
	Name    string
	Version string
}

func DetSemaphoreBytes() []byte {
	d := DetSemaphore{
		Name:    "detergent",
		Version: DetApiVersion,
	}
	b, _ := json.Marshal(d)
	return b
}
