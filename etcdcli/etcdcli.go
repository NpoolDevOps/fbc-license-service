package etcdcli

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/coreos/go-etcd/etcd"
	"golang.org/x/xerrors"
	"os"
)

func Get(key string) ([]byte, error) {
	etcdHost := types.EtcdHost
	env, ok := os.LookupEnv("ETCD_HOST_TEST")
	if ok {
		etcdHost = env
	}

	etcdCli := etcd.NewClient([]string{etcdHost})

	resp, err := etcdCli.Get(fmt.Sprintf("root/%v", key), true, true)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot get '%v' from %v", key, types.EtcdHost)
		return nil, err
	}

	if resp.Node == nil {
		log.Errorf(log.Fields{}, "no response '%v' from %v", key, types.EtcdHost)
		return nil, xerrors.Errorf("empty response from etcd")
	}

	return []byte(resp.Node.Value), nil
}
