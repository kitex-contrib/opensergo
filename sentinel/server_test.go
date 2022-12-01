// Copyright 2021 CloudWeGo Authors.
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
//

package sentinel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sentinel_api "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/stretchr/testify/assert"
)

// HelloImpl implements the last service interface defined in the IDL.
type HelloImpl struct{}

// Echo implements the HelloImpl interface.
func (s *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{Message: req.Message}
	return
}

func TestSentinelServerMiddleware(t *testing.T) {
	bf := func(ctx context.Context, req, resp interface{}, blockErr error) error {
		return errors.New(FakeErrorMsg)
	}
	srv := hello.NewServer(new(HelloImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "hello"}),
		server.WithMiddleware(SentinelServerMiddleware(
			WithBlockFallback(bf),
		)))
	go srv.Run()
	defer srv.Stop()
	time.Sleep(1 * time.Second)

	c, err := hello.NewClient("hello", client.WithHostPorts(":8888"))
	assert.Nil(t, err)

	err = sentinel_api.InitDefault()
	assert.Nil(t, err)
	req := &api.Request{}
	t.Run("success", func(t *testing.T) {
		_, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               "hello:echo",
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)
		_, err := c.Echo(context.TODO(), req)
		assert.Nil(t, err)

		t.Run("second fail", func(t *testing.T) {
			_, err := c.Echo(context.TODO(), req)
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), FakeErrorMsg))
		})
	})
}
