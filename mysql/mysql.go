package fbcmysql

import (
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/xerrors"
	"time"
)

type MysqlConfig struct {
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
	DbName string `json:"db"`
}

type MysqlCli struct {
	config MysqlConfig
	url    string
	db     *gorm.DB
}

func NewMysqlCli(config MysqlConfig) *MysqlCli {
	cli := &MysqlCli{
		config: config,
		url: fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host, config.DbName),
	}

	log.Infof(log.Fields{}, "open mysql db %v", cli.url)
	db, err := gorm.Open("mysql", cli.url)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot open %v: %v", cli.url, err)
		return nil
	}

	log.Infof(log.Fields{}, "successful to create mysql db %v", cli.url)
	db.SingularTable(true)
	cli.db = db

	return cli
}

func (cli *MysqlCli) Delete() {
	cli.db.Close()
}

type UserInfo struct {
	Id           uuid.UUID `gorm:"column:id;primary_key"`
	UserName     string    `gorm:"column:username"`
	ValidateDate time.Time `gorm:"column:validate_date"`
	Quota        int       `gorm:"column:quota"`
	Count        int       `gorm:"column:count"`
	CreateTime   time.Time `gorm:"column:create_time"`
	ModifyTime   time.Time `gorm:"column:modify_time"`
}

func (cli *MysqlCli) QueryUserInfo(user string) (*UserInfo, error) {
	var info UserInfo
	var count int

	cli.db.Where("username = ?", user).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find any value")
	}

	return &info, nil
}

type StatusInfo struct {
	Id       string `gorm:"column:id;primary_key"`
	StatText string `gorm:"column:status_text"`
}

func (cli *MysqlCli) QueryStatusInfo(status string) (*StatusInfo, error) {
	var info StatusInfo
	var count int

	cli.db.Where("status_text = ?", status).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find any value")
	}

	return &info, nil
}

type ClientInfo struct {
	Id         uuid.UUID `gorm:"column:id;primary_key"`
	ClientUser string    `gorm:"column:client_user"`
	ClientSn   string    `gorm:"column:client_sn"`
	Status     string    `gorm:"column:status"`
	CreateTime time.Time `gorm:"column:create_time"`
	ModifyTime time.Time `gorm:"column:modify_time"`
}

func (cli *MysqlCli) InsertClientInfo(info ClientInfo) error {
	_, err := cli.QueryUserInfo(info.ClientUser)
	if err != nil {
		return err
	}

	_, err = cli.QueryStatusInfo(info.Status)
	if err != nil {
		return err
	}

	rc := cli.db.Create(&info)
	return rc.Error
}

func (cli *MysqlCli) QueryClientInfoByClientSn(sn string) (*ClientInfo, error) {
	var info ClientInfo
	var count int

	cli.db.Where("client_sn = ?", sn).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find client")
	}

	return &info, nil
}

func (cli *MysqlCli) QueryClientCount(user string) int {
	var infos []ClientInfo
	var count int

	cli.db.Where("client_user = ?", user).Find(&infos).Count(&count)

	return count
}

func (cli *MysqlCli) QueryClientInfos() []ClientInfo {
	var infos []ClientInfo

	cli.db.Find(&infos)

	return infos
}

func (cli *MysqlCli) QueryClientStatus(id uuid.UUID) *ClientInfo {
	var info ClientInfo
	var count int

	cli.db.Where("id = ?", id).Find(&info).Count(&count)
	if count == 0 {
		return nil
	}

	return &info
}
