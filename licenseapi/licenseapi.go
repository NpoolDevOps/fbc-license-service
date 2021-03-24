package licenseapi

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
)

const licenseDomain = "license.npool.top"

type licenseHostConfig struct {
	Host string `json:"host"`
}

func getLicenseHost() (string, error) {
	var myConfig licenseHostConfig

	resp, err := etcdcli.Get(licenseDomain)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot get %v: %v", licenseDomain, err)
		return "", err
	}

	err = json.Unmarshal([]byte(resp[0]), &myConfig)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse %v: %v", string(resp[0]), err)
		return "", err
	}

	return myConfig.Host, err
}

func ClientInfo(input types.ClientInfoInput) (*types.ClientInfoOutput, error) {
	host, err := getLicenseHost()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", licenseDomain, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.ClientInfoAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.ClientInfoAPI))
	if err != nil {
		log.Errorf(log.Fields{}, "heartbeat error: %v", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, xerrors.Errorf("NON-200 return")
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	output := types.ClientInfoOutput{}
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}
