module github.com/NpoolDevOps/fbc-license-service

go 1.15

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210210082337-af230fd03ce7
	github.com/NpoolDevOps/fbc-auth-service v0.0.0-20210319111238-cce28019e201
	github.com/NpoolRD/http-daemon v0.0.0-20210210091512-241ac31803ef
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/uuid v1.2.0
	github.com/jinzhu/gorm v1.9.16
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20210317182105-75c7a8546eb9 // indirect
	google.golang.org/grpc v1.36.0 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
