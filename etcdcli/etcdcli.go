package etcdcli

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/xerrors"
	"os"
	"time"
)

func Get(key string) ([][]byte, error) {
	os.Setenv("GRPC_GO_REQUIRE_HANDSHAKE", "off")

	etcdHost := types.EtcdHost
	env, ok := os.LookupEnv("ETCD_HOST_TEST")
	if ok {
		etcdHost = env
	}

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdHost},
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := etcdCli.Get(ctx, fmt.Sprintf("root/%v", key))
	if err != nil {
		log.Errorf(log.Fields{}, "cannot get '%v' from %v", key, etcdHost)
		return nil, err
	}

	vals := [][]byte{}
	for _, ev := range resp.Kvs {
		vals = append(vals, ev.Value)
	}

	if len(vals) == 0 {
		return nil, xerrors.Errorf("empty response from etcd")
	}

	return vals, nil
}

type HostConfig struct {
	Host string `json:"host"`
}

func GetHostByDomain(domain string) (string, error) {
	var myConfig HostConfig

	resp, err := Get(domain)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot get %v: %v", domain, err)
		return "", err
	}

	err = json.Unmarshal([]byte(resp[0]), &myConfig)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse %v: %v", string(resp[0]), err)
		return "", err
	}

	return myConfig.Host, err
}
