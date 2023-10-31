module github.com/kitex-contrib/opensergo

go 1.16

require (
	github.com/alibaba/sentinel-golang v1.0.4
	github.com/cloudwego/kitex v0.7.3
	github.com/cloudwego/kitex-examples v0.1.0
	github.com/opensergo/opensergo-go v0.0.0-20221129091737-554d5c0b9105
	github.com/stretchr/testify v1.8.3
	google.golang.org/grpc v1.59.0
)

// TODO: remove this after merging the kitex grpc multi-service feature
replace github.com/cloudwego/kitex => github.com/Marina-Sakai/kitex v0.0.0-20231030012654-522e2fe3f7c7
