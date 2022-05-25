package opensergo

import (
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/opensergo/metainfo"
)

// todo add a example

// NewServer creates a server.Server with the given srvInfo、handler and options.
func NewServer(srvInfo *serviceinfo.ServiceInfo, handler interface{}, opts ...server.Option) server.Server {
	var options []server.Option

	options = append(options, opts...)

	svr := server.NewServer(options...)
	if err := svr.RegisterService(srvInfo, handler); err != nil {
		panic(err)
	}

	// todo define the start-up method
	server.RegisterStartHook(func() {
		metainfo.ReportMetaInfo(svr.GetServiceInfo())
	})
	return svr
}
