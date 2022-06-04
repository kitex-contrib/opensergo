// Copyright 2021 CloudWeGo authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type OpenSergoConfig struct {
	Endpoint string `json:"endpoint"`
}

func GetOpenSergoConfig() (*OpenSergoConfig, error) {
	// refer to https://github.com/opensergo/opensergo-specification/blob/main/specification/en/README.md
	c := OpenSergoConfig{
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
