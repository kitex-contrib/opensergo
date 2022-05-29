package opensergo

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/kitex-contrib/opensergo/util"
	v1 "github.com/opensergo/opensergo-go/proto/service_contract/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type opensergoHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type opensergoMetaReporter struct {
	client          v1.MetadataServiceClient
	refreshInterval time.Duration
	lock            *sync.RWMutex
	reportMetaIns   map[string]*opensergoHeartbeat
}

func NewDefaultMetaReporter() (registry.Registry, error) {
	config, err := util.GetOpensergoConfig()
	if err != nil {
		klog.Errorf("err: %+v", err)
		return nil, err
	}
	return NewMetaReporter(config.Endpoint, util.DefaultRefreshInterval)
}

func NewMetaReporter(endpoint string, refreshInterval time.Duration) (registry.Registry, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Errorf("err: %+v", err)
		return nil, err
	}
	mClient := v1.NewMetadataServiceClient(conn)
	return &opensergoMetaReporter{
		client:          mClient,
		refreshInterval: refreshInterval,
		lock:            &sync.RWMutex{},
		reportMetaIns:   make(map[string]*opensergoHeartbeat),
	}, nil
}

func (o *opensergoMetaReporter) Register(info *registry.Info) error {
	// pack opensergo meta info
	metaReq, err := o.opensergoMeta(info)
	if err != nil {
		klog.Errorf("err:%+v", err)
		return err
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())

	o.lock.RLock()
	_, ok := o.reportMetaIns[instanceKey]
	o.lock.RUnlock()
	if ok {
		return fmt.Errorf("instance{%s} already registered", instanceKey)
	}

	// report opensergo meta info
	if _, err = o.client.ReportMetadata(context.TODO(), metaReq); err != nil {
		return err
	}

	// cron report meta info
	ctx, cancel := context.WithCancel(context.Background())
	go o.refreshMetaInfo(ctx, metaReq)

	o.lock.Lock()
	defer o.lock.Unlock()
	o.reportMetaIns[instanceKey] = &opensergoHeartbeat{
		instanceKey: instanceKey,
		cancel:      cancel,
	}

	return nil
}

func (o *opensergoMetaReporter) Deregister(info *registry.Info) error {
	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())
	o.lock.RLock()
	reportMeta, ok := o.reportMetaIns[instanceKey]
	o.lock.RUnlock()
	if !ok {
		return fmt.Errorf("instance{%s} has not registered", instanceKey)
	}
	o.lock.Lock()
	reportMeta.cancel()
	delete(o.reportMetaIns, instanceKey)
	o.lock.Unlock()
	return nil
}

func (o *opensergoMetaReporter) opensergoMeta(info *registry.Info) (*v1.ReportMetadataRequest, error) {
	serviceDesc := &v1.ServiceDescriptor{
		Name: info.ServiceName,
	}
	streaming := false
	isStreaming, exist := info.ServiceInfo.Extra["streaming"]
	if exist {
		streaming = isStreaming.(bool)
	}
	for methodName, method := range info.ServiceInfo.Methods {
		serviceDesc.Methods = append(serviceDesc.Methods, &v1.MethodDescriptor{
			Name:            info.ServiceInfo.ServiceName + "." + methodName,
			InputTypes:      []string{fmt.Sprintf("%T", method.NewArgs())},
			OutputTypes:     []string{fmt.Sprintf("%T", method.NewResult())},
			ClientStreaming: &streaming,
			ServerStreaming: &streaming,
		})
	}
	serviceContract := v1.ServiceContract{
		Services: []*v1.ServiceDescriptor{serviceDesc},
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return nil, err
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("parse registry info port error: %w", err)
	}
	if host == "" || host == "::" {
		host, err = util.GetLocalIpv4Host()
		if err != nil {
			return nil, fmt.Errorf("parse registry info addr error: %w", err)
		}
	}

	serviceMetadata := v1.ServiceMetadata{
		ServiceContract: &serviceContract,
		Protocols:       []string{info.ServiceInfo.PayloadCodec.String()},
		ListeningAddresses: []*v1.SocketAddress{
			{
				Address:   host,
				PortValue: uint32(p),
			},
		},
	}

	return &v1.ReportMetadataRequest{
		AppName:         info.ServiceName,
		ServiceMetadata: []*v1.ServiceMetadata{&serviceMetadata},
	}, nil
}

func (o *opensergoMetaReporter) refreshMetaInfo(ctx context.Context, metaReq *v1.ReportMetadataRequest) {
	ticker := time.NewTicker(o.refreshInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if _, err := o.client.ReportMetadata(ctx, metaReq); err != nil {
				klog.Errorf("reportMetadata err:%+v", err)
			}
		}
	}
}
