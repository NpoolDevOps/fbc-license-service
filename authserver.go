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
		Location: types.HeartbeatV1API,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.HeartbeatV1Request(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.MyClientsAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.MyClientsRequest(w, req)
		},
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.UpdateAuthAPI,
		Method:   "POST",
		Handler: func(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
			return s.UpdateAuthRequest(w, req)
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
	}

	err = s.redisClient.InsertKeyInfo("session", sessionId,
		fbcredis.SessionInfo{
			MyPubKey:     myPubKey,
			ClientPubKey: input.PublicKey,
		}, 24*100000*time.Hour)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to insert session info: %v", err)
		return nil, err.Error(), -4
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
			Id:          uuid.New(),
			ClientUser:  input.ClientUser,
			ClientSn:    input.ClientSN,
			NetworkType: input.NetworkType,
			Status:      "online",
			CreateTime:  time.Now(),
			ModifyTime:  time.Now(),
		}
		err = s.mysqlClient.InsertClientInfo(*clientInfo)
		if err != nil {
			log.Errorf(log.Fields{}, "fail to insert client info: %v", err)
			return nil, err.Error(), -6
		}
	} else {
		if clientInfo.ClientUser != input.ClientUser {
			return nil, "registered user and client report user is not equal", -7
		}
	}

	clientInfo.NetworkType = input.NetworkType
	s.redisClient.InsertKeyInfo("client", clientInfo.Id, clientInfo, 2*time.Hour)

	return types.ClientLoginOutput{
		ClientUuid: clientInfo.Id,
	}, "", 0
}

func (s *AuthServer) heartbeatRequest(w http.ResponseWriter, req *http.Request) ([]byte, interface{}, string, int) {
	b, _ := ioutil.ReadAll(req.Body)

	var input = types.HeartbeatInput{}
	err := json.Unmarshal(b, &input)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse input parameter: %v [%v]", err, string(b))
		return nil, nil, err.Error(), -1
	}

	sessionInfo, err := s.redisClient.QuerySession(input.SessionId)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to query session: %v", err)
		return nil, nil, err.Error(), -3
	}

	clientInfo, err := s.mysqlClient.QueryClientInfoByClientId(input.ClientUuid)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to find client info: %v", err)
		return nil, nil, err.Error(), -4
	}

	s.redisClient.InsertKeyInfo("client", clientInfo.Id, clientInfo, 2*time.Hour)

	shouldStop := false
	switch clientInfo.Status {
	case fbcmysql.StatusDisable:
		shouldStop = true
	}

	output := types.HeartbeatOutput{
		ShouldStop: shouldStop,
	}

	return []byte(sessionInfo.MyPubKey), output, "", 0
}

func (s *AuthServer) HeartbeatRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	pubKey, output, msg, code := s.heartbeatRequest(w, req)
	if code != 0 {
		return nil, msg, code
	}

	b, _ := json.Marshal(output)
	remoteRsa := crypto.NewRsaCryptoWithParam([]byte(pubKey), nil)
	cipherText, _ := remoteRsa.Encrypt(b)

	return hex.EncodeToString(cipherText), "", 0
}

func (s *AuthServer) HeartbeatV1Request(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	_, output, msg, code := s.heartbeatRequest(w, req)
	return output, msg, code
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
		output.Users = []types.UserInfo{*clientUser}
		output.Clients = s.mysqlClient.QueryClientInfosByUser(clientUser.Username)
	}

	for i, client := range output.Clients {
		expire, err := s.redisClient.QueryClientExpire(client.Id)
		if err != nil {
			output.Clients[i].Status = fbcmysql.StatusDisable
		} else if expire {
			if client.Status == fbcmysql.StatusOnline {
				output.Clients[i].Status = fbcmysql.StatusOffline
			}
		}
	}

	return output, "", 0
}

func (s *AuthServer) UpdateAuthRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	input := types.UpdateAuthInput{}
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

	if !user.SuperUser {
		return nil, "operation not allowed", -5
	}

	usernameInfo, err := authapi.UsernameInfo(authtypes.UsernameInfoInput{
		AuthCode: input.AuthCode,
		Username: input.Username,
	})
	if err != nil {
		return nil, err.Error(), -6
	}

	clientUser, err := s.mysqlClient.QueryUserInfoById(usernameInfo.Id)
	if err != nil {
		clientUser = &types.UserInfo{
			Id:         usernameInfo.Id,
			Username:   input.Username,
			CreateTime: time.Now(),
		}
	}

	clientUser.Quota = input.Quota
	clientUser.ModifyTime = time.Now()
	clientUser.ValidateDate = time.Now().AddDate(0, 0, input.ValidateDate)

	err = s.mysqlClient.UpdateAuth(*clientUser)
	if err != nil {
		return nil, err.Error(), -6
	}

	return nil, "", 0
}
