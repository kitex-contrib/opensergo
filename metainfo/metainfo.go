package metainfo

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/kitex-contrib/opensergo/config"
	service_contract_v1 "github.com/opensergo/opensergo-go/proto/service_contract/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

func ReportMetaInfo(srvInfo *serviceinfo.ServiceInfo) {
	//var services []*service_contract_v1.ServiceDescriptor
	var serviceDesc = &service_contract_v1.ServiceDescriptor{}
	serviceDesc.Name = srvInfo.ServiceName
	for methodName, _ := range srvInfo.Methods {
		serviceDesc.Methods = append(serviceDesc.Methods, &service_contract_v1.MethodDescriptor{
			Name:        srvInfo.ServiceName + "." + methodName,
			InputTypes:  nil, // todo add inputType and outputType
			OutputTypes: nil,
		})
	}

	serviceContract := service_contract_v1.ServiceContract{
		Services: []*service_contract_v1.ServiceDescriptor{serviceDesc},
	}

	serviceMetadata := service_contract_v1.ServiceMetadata{
		ServiceContract:    &serviceContract,
		Protocols:          []string{srvInfo.PayloadCodec.String()},
		ListeningAddresses: nil, // todo add ListeningAddresses
	}

	timeoutCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(timeoutCtx, config.OpenSergoEndpoint(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorf("err: %v", err)
		return
	}
	mClient := service_contract_v1.NewMetadataServiceClient(conn)
	if _, err = mClient.ReportMetadata(context.TODO(), &service_contract_v1.ReportMetadataRequest{
		AppName:         srvInfo.ServiceName,
		ServiceMetadata: []*service_contract_v1.ServiceMetadata{&serviceMetadata},
	}); err != nil {
		klog.Errorf("err: %v", err)
		return
	}
	return
}
