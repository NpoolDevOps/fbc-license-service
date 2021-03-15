package main

import (
	"encoding/json"
	"fbclicenseserver/crypto"
	fbclib "fbclicenseserver/library"
	fbcmysql "fbclicenseserver/mysql"
	fbcredis "fbclicenseserver/redis"
	types "fbclicenseserver/types"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolRD/http-daemon"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
)

type PairedCrypto struct {
	RemoteRsa *crypto.RsaCrypto
	LocalRsa  *crypto.RsaCrypto
}

type AuthServerConfig struct {
	RedisCfg fbcredis.RedisConfig `json:"redis"`
	MysqlCfg fbcmysql.MysqlConfig `json:"mysql"`
	Port     int                  `json:"port"`
}

type AuthServer struct {
	config       AuthServerConfig
	authText     string
	redisClient  *fbcredis.RedisCli
	mysqlClient  *fbcmysql.MysqlCli
	clientCrypto map[uuid.UUID]PairedCrypto
}

func NewAuthServer(configFile string) *AuthServer {
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil
	}

	config := AuthServerConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return nil
	}

	redisCli := fbcredis.NewRedisCli(config.RedisCfg)
	if redisCli == nil {
		return nil
	}

	mysqlCli := fbcmysql.NewMysqlCli(config.MysqlCfg)
	if mysqlCli == nil {
		return nil
	}

	server := &AuthServer{
		config:       config,
		authText:     fbclib.FBCAuthText,
		redisClient:  redisCli,
		mysqlClient:  mysqlCli,
		clientCrypto: make(map[uuid.UUID]PairedCrypto),
	}

	return server
}

func (s *AuthServer) Run() error {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/api/v0/client/exchange_key",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.ExchangeKeyRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/api/v0/client/startup",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.StartUpRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/api/v0/client/heartbeat",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.HeartbeatRequest(w, req)
		},
	})

	httpdaemon.Run(s.config.Port)
	return nil
}

func (s *AuthServer) ExchangeKeyRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to process exchange key: %v", err)
		return nil, err.Error(), -1
	}

	var exchangeKeyInput types.ExchangeKeyInput
	err = json.Unmarshal(b, &exchangeKeyInput)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse input paramster: %v [%v]", err, string(b))
		return nil, err.Error(), -2
	}

	sessionId := uuid.New()

	s.clientCrypto[sessionId] = PairedCrypto{
		RemoteRsa: crypto.NewRsaCryptoWithParam([]byte(exchangeKeyInput.PublicKey), nil),
		LocalRsa:  crypto.NewRsaCrypto(1024),
	}

	return types.ExchangeKeyOutput{
		PublicKey: string(s.clientCrypto[sessionId].LocalRsa.GetPubkey()),
		SessionId: sessionId,
	}, "", 0
}

func (s *AuthServer) StartUpRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}

func (s *AuthServer) HeartbeatRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}
