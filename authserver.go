package main

import (
	"encoding/hex"
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	authapi "github.com/NpoolDevOps/fbc-auth-service/authapi"
	authtypes "github.com/NpoolDevOps/fbc-auth-service/types"
	"github.com/NpoolDevOps/fbc-license-service/crypto"
	fbclib "github.com/NpoolDevOps/fbc-license-service/library"
	fbcmysql "github.com/NpoolDevOps/fbc-license-service/mysql"
	fbcredis "github.com/NpoolDevOps/fbc-license-service/redis"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/NpoolRD/http-daemon"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"time"
)

type AuthServerConfig struct {
	RedisCfg fbcredis.RedisConfig `json:"redis"`
	MysqlCfg fbcmysql.MysqlConfig `json:"mysql"`
	Port     int                  `json:"port"`
}

type AuthServer struct {
	config      AuthServerConfig
	authText    string
	redisClient *fbcredis.RedisCli
	mysqlClient *fbcmysql.MysqlCli
}

func NewAuthServer(configFile string) *AuthServer {
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read file %v: %v", configFile, err)
		return nil
	}

	config := AuthServerConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse file %v: %v", configFile, err)
		return nil
	}

	log.Infof(log.Fields{}, "create redis cli: %v", config.RedisCfg)
	redisCli := fbcredis.NewRedisCli(config.RedisCfg)
	if redisCli == nil {
		log.Errorf(log.Fields{}, "cannot create redis client %v: %v", config.RedisCfg, err)
		return nil
	}

	log.Infof(log.Fields{}, "create mysql cli: %v", config.MysqlCfg)
	mysqlCli := fbcmysql.NewMysqlCli(config.MysqlCfg)
	if mysqlCli == nil {
		log.Errorf(log.Fields{}, "cannot create mysql client %v: %v", config.MysqlCfg, err)
		return nil
	}

	server := &AuthServer{
		config:      config,
		authText:    fbclib.FBCAuthText,
		redisClient: redisCli,
		mysqlClient: mysqlCli,
	}

	log.Infof(log.Fields{}, "successful to create auth server")

	return server
}

func (s *AuthServer) Run() error {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.ExchangeKeyAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.ExchangeKeyRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.LoginAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.LoginRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.HeartbeatAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.HeartbeatRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.MyClientsAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.MyClientsRequest(w, req)
		},
	})

	log.Infof(log.Fields{}, "start http daemon at %v", s.config.Port)
	httpdaemon.Run(s.config.Port)
	return nil
}

func (s *AuthServer) ExchangeKeyRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to process exchange key: %v", err)
		return nil, err.Error(), -1
	}

	var input types.ExchangeKeyInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse input parameter: %v [%v]", err, string(b))
		return nil, err.Error(), -2
	}

	if input.Spec == "" {
		log.Errorf(log.Fields{}, "device spec is must")
		return nil, "device spec is must", -3
	}

	var sessionId uuid.UUID
	sessionExist := false

	device, err := s.redisClient.QueryDevice(input.Spec)
	if err == nil {
		sessionId = device.SessionId
		_, err := s.redisClient.QuerySession(sessionId)
		if err != nil {
			return nil, err.Error(), -4
		}
		sessionExist = true
	}

	var localRsa *crypto.RsaCrypto
	var myPubKey string

	if !sessionExist {
		sessionId = uuid.New()
		localRsa = crypto.NewRsaCrypto(1024)
		myPubKey = string(localRsa.GetPubkey())

		err = s.redisClient.InsertKeyInfo("session", sessionId,
			fbcredis.SessionInfo{
				MyPubKey:     myPubKey,
				ClientPubKey: input.PublicKey,
			}, 24*100000*time.Hour)
		if err != nil {
			log.Errorf(log.Fields{}, "fail to insert session info: %v", err)
			return nil, err.Error(), -4
		}
	}

	return types.ExchangeKeyOutput{
		PublicKey: myPubKey,
		SessionId: sessionId,
	}, "", 0
}

func (s *AuthServer) LoginRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, _ := ioutil.ReadAll(req.Body)

	var input = types.ClientLoginInput{}
	err := json.Unmarshal(b, &input)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse input parameter: %v [%v]", err, string(b))
		return nil, err.Error(), -1
	}

	log.Infof(log.Fields{}, "login request from %v / %v", input.ClientUser, input.ClientPasswd)
	myAppId := uuid.MustParse("00000001-0001-0001-0001-000000000001")
	_, err = authapi.Login(authtypes.UserLoginInput{
		Username: input.ClientUser,
		Password: input.ClientPasswd,
		AppId:    myAppId,
	})
	if err != nil {
		log.Errorf(log.Fields{}, "fail to login: %v", err)
		return nil, err.Error(), -2
	}

	_, err = s.redisClient.QuerySession(input.SessionId)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to query session: %v", err)
		return nil, err.Error(), -4
	}

	_, err = s.mysqlClient.QueryUserInfoByUsername(input.ClientUser)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to find client user: %v", err)
		return nil, err.Error(), -5
	}

	clientInfo, err := s.mysqlClient.QueryClientInfoByClientSn(input.ClientSN)
	if err != nil {
		clientInfo = &types.ClientInfo{
			Id:         uuid.New(),
			ClientUser: input.ClientUser,
			ClientSn:   input.ClientSN,
			Status:     "online",
			CreateTime: time.Now(),
			ModifyTime: time.Now(),
		}
		err = s.mysqlClient.InsertClientInfo(*clientInfo)
		if err != nil {
			log.Errorf(log.Fields{}, "fail to insert client info: %v", err)
			return nil, err.Error(), -6
		}
	}

	s.redisClient.InsertKeyInfo("client", clientInfo.Id, clientInfo, 24*100000*time.Hour)

	return types.ClientLoginOutput{
		ClientUuid: clientInfo.Id,
	}, "", 0
}

func (s *AuthServer) HeartbeatRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, _ := ioutil.ReadAll(req.Body)

	var input = types.HeartbeatInput{}
	err := json.Unmarshal(b, &input)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse input parameter: %v [%v]", err, string(b))
		return nil, err.Error(), -1
	}

	sessionInfo, err := s.redisClient.QuerySession(input.SessionId)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to query session: %v", err)
		return nil, err.Error(), -3
	}

	clientInfo, err := s.mysqlClient.QueryClientInfoByClientId(input.ClientUuid)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to find client info: %v", err)
		return nil, err.Error(), -4
	}

	shouldStop := false
	switch clientInfo.Status {
	case fbcmysql.StatusDisable:
		shouldStop = true
	}

	output := types.HeartbeatOutput{
		ShouldStop: shouldStop,
	}

	b, _ = json.Marshal(output)
	remoteRsa := crypto.NewRsaCryptoWithParam([]byte(sessionInfo.MyPubKey), nil)
	cipherText, _ := remoteRsa.Encrypt(b)

	return hex.EncodeToString(cipherText), "", 0
}

func (s *AuthServer) MyClientsRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	input := types.MyClientsInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -2
	}

	if input.AuthCode == "" {
		return nil, "auth code is must", -3
	}

	user, err := authapi.UserInfo(authtypes.UserInfoInput{
		AuthCode: input.AuthCode,
	})
	if err != nil {
		return nil, err.Error(), -4
	}

	clientUser, err := s.mysqlClient.QueryUserInfoById(user.Id)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to find client user: %v", err)
		return nil, err.Error(), -5
	}

	output := types.MyClientsOutput{
		SuperUser:   user.SuperUser,
		VisitorOnly: user.VisitorOnly,
	}
	if user.SuperUser {
		output.Clients = s.mysqlClient.QueryClientInfos()
		output.Users = s.mysqlClient.QueryUserInfos()
	} else {
		output.Clients = s.mysqlClient.QueryClientInfosByUser(clientUser.Username)
	}

	return output, "", 0
}
