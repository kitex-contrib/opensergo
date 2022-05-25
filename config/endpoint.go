package config

import (
	"encoding/json"
	"github.com/cloudwego/kitex/pkg/klog"
	"io/ioutil"
	"os"
)

const (
	OpensergoBootstrapConfig = "OPENSERGO_BOOTSTRAP_CONFIG"
	OpensergoBootstrap       = "OPENSERGO_BOOTSTRAP"
)

type OpenSergoConfig struct {
	Endpoint string `json:"config"`
}

func OpenSergoEndpoint() string {
	var err error
	configStr := os.Getenv(OpensergoBootstrapConfig)
	configBytes := []byte(configStr)
	if configStr == "" {
		configPath := os.Getenv(OpensergoBootstrap)
		configBytes, err = ioutil.ReadFile(configPath)
		if err != nil {
			klog.Errorf("err: %v", err)
		}
	}
	config := OpenSergoConfig{}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		klog.Errorf("err: %v", err)
	}
	return config.Endpoint
}
