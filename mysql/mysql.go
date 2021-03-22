package fbcmysql

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/xerrors"
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

const (
	StatusOnline      = "online"
	StatusOffline     = "offline"
	StatusMaintaining = "maintaining"
	StatusDisable     = "disable"
)

func NewMysqlCli(config MysqlConfig) *MysqlCli {
	cli := &MysqlCli{
		config: config,
		url: fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host, config.DbName),
	}

	var myConfig MysqlConfig

	resp, err := etcdcli.Get(config.Host)
	if err == nil {
		err = json.Unmarshal(resp[0], &myConfig)
		if err == nil {
			myConfig.DbName = config.DbName
			cli = &MysqlCli{
				config: myConfig,
				url: fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
					myConfig.User, myConfig.Passwd, myConfig.Host, myConfig.DbName),
			}
		}
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

func (cli *MysqlCli) QueryUserInfoByUsername(user string) (*types.UserInfo, error) {
	var info types.UserInfo
	var count int

	cli.db.Where("username = ?", user).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find any value")
	}

	return &info, nil
}

func (cli *MysqlCli) QueryUserInfoById(uid uuid.UUID) (*types.UserInfo, error) {
	var info types.UserInfo
	var count int

	cli.db.Where("id = ?", uid).Find(&info).Count(&count)
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

func (cli *MysqlCli) InsertClientInfo(info types.ClientInfo) error {
	_, err := cli.QueryUserInfoByUsername(info.ClientUser)
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

func (cli *MysqlCli) QueryClientInfoByClientSn(sn string) (*types.ClientInfo, error) {
	var info types.ClientInfo
	var count int

	cli.db.Where("client_sn = ?", sn).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find client")
	}

	return &info, nil
}

func (cli *MysqlCli) QueryClientInfoByClientId(id uuid.UUID) (*types.ClientInfo, error) {
	var info types.ClientInfo
	var count int

	cli.db.Where("id = ?", id).Find(&info).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("cannot find client")
	}

	return &info, nil
}

func (cli *MysqlCli) QueryClientCount(user string) int {
	var infos []types.ClientInfo
	var count int

	cli.db.Where("client_user = ?", user).Find(&infos).Count(&count)

	return count
}

func (cli *MysqlCli) QueryUserInfos() []types.UserInfo {
	var infos []types.UserInfo

	cli.db.Find(&infos)

	return infos
}

func (cli *MysqlCli) QueryClientInfos() []types.ClientInfo {
	var infos []types.ClientInfo

	cli.db.Find(&infos)

	return infos
}

func (cli *MysqlCli) QueryClientInfosByUser(username string) []types.ClientInfo {
	var infos []types.ClientInfo

	var count = 0

	cli.db.Where("client_user = ?", username).Find(&infos).Count(&count)
	if count == 0 {
		return nil
	}

	return infos
}

func (cli *MysqlCli) QueryClientStatus(id uuid.UUID) *types.ClientInfo {
	var info types.ClientInfo
	var count int

	cli.db.Where("id = ?", id).Find(&info).Count(&count)
	if count == 0 {
		return nil
	}

	return &info
}
