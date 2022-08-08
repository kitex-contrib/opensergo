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

package sentinel

import (
	"context"
	"errors"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/stretchr/testify/assert"
)

const FakeErrorMsg = "fake error for testing"

func TestSentinelClientMiddleware(t *testing.T) {
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	c, err := hello.NewClient("hello",
		client.WithMiddleware(SentinelClientMiddleware(
			WithBlockFallback(bf))))
	if err != nil {
		t.Fatal(err)
	}
	err = sentinel.InitDefault()
	if err != nil {
		t.Fatal(err)
	}
	req := &api.Request{}
	t.Run("success", func(t *testing.T) {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               "hello:echo",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		_, err = c.Echo(context.Background(), req)
		assert.NotNil(t, err)
		assert.NotEqual(t, FakeErrorMsg, err.Error())
		t.Run("second fail", func(t *testing.T) {
			_, err = c.Echo(context.Background(), req)
			assert.NotNil(t, err)
			assert.Equal(t, FakeErrorMsg, err.Error())
		})
	})
}
