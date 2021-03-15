package main

import (
	"encoding/json"
	fbclib "fbclicenseserver/library"
	"io/ioutil"
)

type AuthServerConfig struct {
	RedisCfg RedisConfig `json:"redis"`
	MysqlCfg MysqlConfig `json:"mysql"`
}

type AuthServer struct {
	config      AuthServerConfig
	authText    string
	redisClient *RedisCli
	mysqlClient *MysqlCli
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

	server := &AuthServer{
		config:      config,
		authText:    fbclib.FBCAuthText,
		redisClient: NewRedisCli(config.RedisCfg),
		mysqlClient: NewMysqlCli(config.MysqlCfg),
	}

	return server
}

func (s *AuthServer) Run() error {
	return nil
}
