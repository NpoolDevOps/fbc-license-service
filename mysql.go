package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type MysqlConfig struct {
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
}

type MysqlCli struct {
	config MysqlConfig
	url    string
	db     *gorm.DB
}

func NewMysqlCli(config MysqlConfig) *MysqlCli {
	cli := &MysqlCli{
		config: config,
		url: fmt.Sprintf("%v:%v@tcp(%v)/fbc-license-db?charset=utf8&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host),
	}

	db, err := gorm.Open("mysql", cli.url)
	if err != nil {
		return nil
	}

	db.SingularTable(true)
	cli.db = db

	return cli
}

func (cli *MysqlCli) Delete() {
	cli.db.Close()
}

type UserInfo struct {
	Id           string    `gorm:"column:id;primary_key"`
	UserName     string    `gorm:"column:user_name"`
	ValidateDate time.Time `gorm:"column:validate_date"`
	Volume       int       `gorm:"column:volume"`
	Sn           string    `gorm:"column:sn"`
	CreateTime   time.Time `gorm:"column:create_time"`
	ModifyTime   time.Time `gorm:"column:modify_time"`
}

func (cli *MysqlCli) QueryUserInfoBySn(db *gorm.DB, softwareSn string) *UserInfo {
	var userInfo UserInfo
	var count int

	db.Where("sn = ?", softwareSn).Find(&userInfo).Count(&count)
	if count == 0 {
		return nil
	}

	return &userInfo
}

type ClientInfo struct {
	Id           string    `gorm:"column:id;primary_key"`
	ClientSn     string    `gorm:"column:software_sn"`
	SystemSn     string    `gorm:"column:system_sn"`
	Status       int       `gorm:"column:status"`
	DevopsStatus int       `gorm:"column:devops_status"`
	CreateTime   time.Time `gorm:"column:create_time"`
	ModifyTime   time.Time `gorm:"column:modify_time"`
}

func (cli *MysqlCli) InsertClientInfo(info ClientInfo) {
	cli.db.Create(&info)
}

func (cli *MysqlCli) QueryClientInfoBySystemSn(sn string) *ClientInfo {
	var info ClientInfo
	var count int

	cli.db.Where("system_sn = ?", sn).Find(&info).Count(&count)
	if count == 0 {
		return nil
	}

	return &info
}

func (cli *MysqlCli) GetClientCount(sn string) int {
	var infos []ClientInfo
	var count int

	cli.db.Where("software_sn = ?", sn).Find(&infos).Count(&count)

	return count
}

func (cli *MysqlCli) QueryClientInfos() []ClientInfo {
	var infos []ClientInfo

	cli.db.Find(&infos)

	return infos
}

func (cli *MysqlCli) GetSoftwareDevopsStatus(uuid string) *ClientInfo {
	var info ClientInfo
	var count int

	cli.db.Where("id=?", uuid).Find(&info).Count(&count)
	if count == 0 {
		return nil
	}

	return &info
}
