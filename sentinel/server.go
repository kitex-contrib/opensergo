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

package sentinel

import (
	"context"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/cloudwego/kitex/pkg/endpoint"
)

// SentinelServerMiddleware returns new server.Middleware
// Default resource name is {service's name}:{method}
// Default block fallback is returning blockError
// Define your own behavior by setting serverOptions
func SentinelServerMiddleware(opts ...Option) func(endpoint.Endpoint) endpoint.Endpoint {
	options := newOptions(opts)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			resourceName := options.ResourceExtract(ctx, req, resp)
			entry, blockErr := api.Entry(
				resourceName,
				api.WithResourceType(base.ResTypeRPC),
				api.WithTrafficType(base.Inbound),
			)
			if blockErr != nil {
				return options.BlockFallback(ctx, req, resp, blockErr)
			}
			defer entry.Exit()
			err := next(ctx, req, resp)
			if err != nil {
				api.TraceError(entry, err)
			}
			return err
		}
	}
}
