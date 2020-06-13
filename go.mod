module github.com/xunull/etcd-manage

go 1.14

require (
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gin-gonic/autotls v0.0.0-20200518075542-45033372a9ad
	github.com/gin-gonic/gin v1.6.3
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1
	go.uber.org/zap v1.15.0
	google.golang.org/genproto v0.0.0-20200612171551-7676ae05be11 // indirect
	google.golang.org/grpc v1.29.1 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0