package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

const (
	DefaultRefreshInterval = 30 * time.Second
)

type OpensergoConfig struct {
	Endpoint string `json:"endpoint"`
}

func GetOpensergoConfig() (*OpensergoConfig, error) {
	// https://github.com/opensergo/opensergo-specification/blob/main/specification/en/README.md
	c := OpensergoConfig{
		Endpoint: os.Getenv("OPENSERGO_ENDPOINT"),
	}
	if v := os.Getenv("OPENSERGO_BOOTSTRAP"); v != "" {
		if err := json.Unmarshal([]byte(v), &c); err != nil {
			return nil, err
		}
	}
	if v := os.Getenv("OPENSERGO_BOOTSTRAP_CONFIG"); v != "" {
		b, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}
