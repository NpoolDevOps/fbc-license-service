package main

import (
	"fmt"
)

type MysqlConfig struct {
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type MysqlCli struct {
	config MysqlConfig
	url    string
}

func NewMysqlCli(config MysqlConfig) *MysqlCli {
	cli := &MysqlCli{
		config: config,
		url: fmt.Sprintf("%v:%v@tcp(%v)/fbc-license-db?charset=utf8&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host),
	}
	return cli
}
