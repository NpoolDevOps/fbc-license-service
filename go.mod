module github.com/NpoolDevOps/fbc-license-service

go 1.15

require (
	github.com/EntropyPool/entropy-logger v0.0.0-20210210082337-af230fd03ce7
	github.com/NpoolDevOps/fbc-auth-service v0.0.0-20210407152903-61cdde5f2787
	github.com/NpoolRD/http-daemon v0.0.0-20210210091512-241ac31803ef
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/google/uuid v1.2.0
	github.com/jinzhu/gorm v1.9.16
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
