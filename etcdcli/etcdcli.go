package etcdcli

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/coreos/go-etcd/etcd"
	"golang.org/x/xerrors"
)

func Get(key string) ([]byte, error) {
	etcdCli := etcd.NewClient([]string{types.EtcdHost})

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
