// Copyright 2021 CloudWeGo Authors
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

package metainfo

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	v1 "github.com/opensergo/opensergo-go/pkg/proto/service_contract/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kitex-contrib/opensergo/util"
)

var protocolMap = map[serviceinfo.PayloadCodec]string{
	serviceinfo.Thrift:   "thrift",
	serviceinfo.Protobuf: "grpc",
}

type OpenSergoMetaReporter struct {
	enable bool
	client v1.MetadataServiceClient
}

// NewDefaultMetaReporter create a default meta info reporter
func NewDefaultMetaReporter() (*OpenSergoMetaReporter, error) {
	c, err := util.GetOpenSergoConfig()
	if err != nil {
		klog.Errorf("err: %+v", err)
		return nil, err
	}
	if c == nil {
		klog.Warn(util.ErrConfigNotFound.Error())
		return &OpenSergoMetaReporter{}, nil
	}
	klog.Infof("get config success,config=%+v", c)
	return NewMetaReporter(c)
}

// NewMetaReporter create a meta info reporter
func NewMetaReporter(c *util.OpenSergoConfig) (*OpenSergoMetaReporter, error) {
	if c == nil {
		klog.Warn(util.ErrConfigNotFound.Error())
		return &OpenSergoMetaReporter{}, nil
	}
	conn, err := grpc.Dial(c.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorf("err: %+v", err)
		return nil, err
	}
	mClient := v1.NewMetadataServiceClient(conn)
	return &OpenSergoMetaReporter{
		client: mClient,
		enable: true,
	}, nil
}

// ReportMetaInfo report meta info to opensergo
func (o *OpenSergoMetaReporter) ReportMetaInfo(srvInfo *serviceinfo.ServiceInfo) error {
	if !o.enable {
		klog.Warn(util.ErrConfigNotFound.Error())
		return util.ErrConfigNotFound
	}
	metaReq, err := o.openSergoMetaReq(srvInfo)
	if err != nil {
		klog.Errorf("err:%+v", err)
		return err
	}
	if _, err = o.client.ReportMetadata(context.TODO(), metaReq); err != nil {
		klog.Errorf("err:%+v", err)
		return err
	}
	return nil
}

func (o *OpenSergoMetaReporter) openSergoMetaReq(srvInfo *serviceinfo.ServiceInfo) (*v1.ReportMetadataRequest, error) {
	serviceDesc := &v1.ServiceDescriptor{
		Name: srvInfo.ServiceName,
	}
	streaming := false
	isStreaming, exist := srvInfo.Extra["streaming"].(bool)
	if exist {
		streaming = isStreaming
	}
	for methodName, method := range srvInfo.Methods {
		serviceDesc.Methods = append(serviceDesc.Methods, &v1.MethodDescriptor{
			Name:            srvInfo.GetPackageName() + "." + methodName,
			InputTypes:      []string{util.ParamTypeName(method.NewArgs())},
			OutputTypes:     []string{util.ParamTypeName(method.NewResult())},
			ClientStreaming: &streaming,
			ServerStreaming: &streaming,
		})
	}

	serviceMetadata := v1.ServiceMetadata{
		ServiceContract: &v1.ServiceContract{
			Services: []*v1.ServiceDescriptor{serviceDesc},
		},
	}

	addr, ok := srvInfo.Extra["address"].(net.Addr)
	if ok {
		host, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			return nil, err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("parse port error: %+v", err)
		}
		if host == "" || host == "::" {
			host, err = util.GetLocalIpv4Host()
			if err != nil {
				return nil, fmt.Errorf("parse host error: %+v", err)
			}
		}
		serviceMetadata.ListeningAddresses = []*v1.SocketAddress{
			{
				Address:   host,
				PortValue: uint32(p),
			},
		}
	}

	if protocol, ok := protocolMap[srvInfo.PayloadCodec]; ok {
		serviceMetadata.Protocols = []string{protocol}
	}

	return &v1.ReportMetadataRequest{
		AppName:         srvInfo.ServiceName,
		ServiceMetadata: []*v1.ServiceMetadata{&serviceMetadata},
	}, nil
}
