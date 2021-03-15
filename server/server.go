package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/go-basic/uuid"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"guard_server/http_daemon"
	"guard_server/msg_struct"
	"guard_server/mysql_orm"
	"guard_server/redis_orm"
	"guard_server/rsa_crypto"
	"guard_server/utils"
	"net/http"
	"time"
)

var redisKeyPrefix = "softwareguardserver_" // prefix + software_info_id

type GuardServer struct {
	rsaObjMap       map[string]*msg_struct.RsaPair
	authText        string
	redisAddr       string
	dbType          string
	dbUrl           string
	redisTtl        int
	redisClient     *redis.Client
	mysqlClient     *gorm.DB
	softwareInfoMap map[string]*redis_orm.RedisSoftwareInfo
}

func NewGuardServer(config map[string]interface{}) *GuardServer {

	rsaObjMap := make(map[string]*msg_struct.RsaPair)
	softwareInfomap := make(map[string]*redis_orm.RedisSoftwareInfo)
	fmt.Printf("%#v\n", config)
	return &GuardServer{
		rsaObjMap:       rsaObjMap,
		authText:        config["authText"].(string),
		redisAddr:       config["redisAddr"].(string),
		dbType:          config["dbType"].(string),
		dbUrl:           config["dbUrl"].(string),
		redisTtl:        config["redisTtl"].(int),
		softwareInfoMap: softwareInfomap,
	}
}

func (self *GuardServer) RegisterHandler() {

	http_daemon.RegisterRouter(http_daemon.HttpRouter{
		Location: "/api_client/exchange_key",
		Handler:  self.ExchangeKeyRequest,
	})

	http_daemon.RegisterRouter(http_daemon.HttpRouter{
		Location: "/api_client/startup",
		Handler:  self.StartUpRequest,
	})

	http_daemon.RegisterRouter(http_daemon.HttpRouter{
		Location: "/api_client/heartbeat",
		Handler:  self.HeartbeatRequest,
	})

}

/*functions used for mysql && redis orm*/
func (self *GuardServer) newRedisClient() (*redis.Client, error) {
	return redis_orm.NewRedisClient(self.redisAddr)
}

func (self *GuardServer) closeRedisClient(redisClient *redis.Client) {
	redisClient.Close()
}

func (self *GuardServer) newMysqlClient() (*gorm.DB, error) {

	cli, err := mysql_orm.NewDbOrm(self.dbType, self.dbUrl)
	if err != nil {
		log.Errorf(log.Fields{}, "Create Mysql Client Failed %v / %v [%v]", self.dbType, self.dbUrl, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "Create Mysql Client success %v / %v", self.dbType, self.dbUrl)

	return cli, nil
}

func (self *GuardServer) closeMysqlClient(cli *gorm.DB) {
	mysql_orm.DbOrmClose(cli)
}

func (self *GuardServer) Boot() {

	db, err := self.newMysqlClient()
	if err != nil {
		log.Errorf(log.Fields{}, "Boot newMysqlclient Failed [%v]", err)
		return
	}
	defer self.closeMysqlClient(db)

	redisDb, err := self.newRedisClient()
	if err != nil {
		log.Errorf(log.Fields{}, "Boot newRedisClient Failed [%v]", err)
		return
	}
	defer self.closeRedisClient(redisDb)

	infos := mysql_orm.QuerySoftwareInfos(db)
	for index := range infos {
		info := infos[index]
		key := redisKeyPrefix + info.Id
		redisInfo := redis_orm.RedisQuerysoftwareInfo(key, redisDb)
		if redisInfo != nil {
			self.softwareInfoMap[redisInfo.Id] = redisInfo
		}
	}

}

func (self *GuardServer) ExchangeKeyRequest(w http.ResponseWriter, req *http.Request) (interface{}, error, int) {

	body, err := utils.ReadRequestBody(req)
	if err != nil {
		log.Errorf(log.Fields{}, "ExchangeKeyRequest ReadRequestBody Failed [%v]", err)
		return nil, err, -1
	}

	pubkey := body.(map[string]interface{})["public_key"]
	remoteRsaObj := rsa_crypto.NewRsaCryptoWithParam([]byte(pubkey.(string)), nil)
	localRsaObj := rsa_crypto.NewRsaCrypto(1024)
	sessionId := uuid.New()
	rsaPair := &msg_struct.RsaPair{
		RemoteRsa: remoteRsaObj,
		LocalRsa:  localRsaObj,
	}
	self.rsaObjMap[sessionId] = rsaPair

	response := make(map[string]interface{})
	response["sessionId"] = sessionId
	response["public_key"] = string(localRsaObj.GetPubkey())

	return response, nil, 0
}

func (self *GuardServer) StartUpRequest(w http.ResponseWriter, req *http.Request) (interface{}, error, int) {

	body, err := utils.ReadRequestBody(req)
	if err != nil {
		log.Errorf(log.Fields{}, "StartUpRequest ReadRequestBody Failed [%v]", err)
		return nil, err, -1
	}

	fmt.Printf("%#v\n", body)
	sessionId := body.(map[string]interface{})["sessionId"].(string)
	data := body.(map[string]interface{})["data"].(string)
	hData, _ := hex.DecodeString(data)
	deData, _ := self.rsaObjMap[sessionId].LocalRsa.Decrypt([]byte(hData))
	var param msg_struct.StartUpReq
	err = json.Unmarshal(deData, &param)
	if err != nil {
		// log.Print(err)
	}
	fmt.Printf("param: %+v\n", param)

	db, err := self.newMysqlClient()
	if err != nil {
		log.Errorf(log.Fields{}, "StartUpRequest newMysqlClient Failed [%v]", err)
		return nil, err, -1
	}
	defer self.closeMysqlClient(db)
	redisDb, err := self.newRedisClient()
	if err != nil {
		log.Errorf(log.Fields{}, "StartUpRequest newRedisClient Failed [%v]", err)
		return nil, err, -1
	}
	defer self.closeRedisClient(redisDb)

	response := make(map[string]interface{})
	response["authText"] = self.authText
	response["sessionId"] = sessionId
	response["startUp"] = false
	dbUserInfo := mysql_orm.QueryUserInfoBySn(db, param.ClientSn)
	queryInfo := mysql_orm.QuerySoftwareInfoBySystemSn(db, param.SystemSn)
	if dbUserInfo != nil {
		if queryInfo != nil {
			redisnewInfo := redis_orm.RedisSoftwareInfo{
				Id:         queryInfo.Id,
				SessionId:  sessionId,
				StartUpreq: param,
				RsaPair:    *self.rsaObjMap[sessionId],
			}
			self.softwareInfoMap[queryInfo.Id] = &redisnewInfo
			redis_orm.RedisInsertNewInfo(redisnewInfo, redisDb, self.redisTtl)
			response["startUp"] = true
			response["softwareUuid"] = queryInfo.Id
		} else {
			// count := mysql_orm.GetSoftwareCount(self.mysqlClient, param.ClientSn)
			// if count < dbUserInfo.Volume{
			newInfo := mysql_orm.SoftwareInfo{
				Id:         uuid.New(),
				SoftwareSn: param.ClientSn,
				SystemSn:   param.SystemSn,
				Status:     STATUS_UP,
				CreateTime: time.Now(),
				ModifyTime: time.Now(),
			}
			mysql_orm.InsertSoftwareInfo(db, newInfo)

			redisnewInfo := redis_orm.RedisSoftwareInfo{
				Id:         newInfo.Id,
				SessionId:  sessionId,
				StartUpreq: param,
				RsaPair:    *self.rsaObjMap[sessionId],
			}
			self.softwareInfoMap[redisnewInfo.Id] = &redisnewInfo
			redis_orm.RedisInsertNewInfo(redisnewInfo, redisDb, self.redisTtl)
			response["startUp"] = true
			response["softwareUuid"] = newInfo.Id
			//}
		}
	}
	fmt.Println(response)
	jresponse, _ := json.Marshal(response)
	ciphertext, _ := self.rsaObjMap[sessionId].RemoteRsa.Encrypt(jresponse)
	return hex.EncodeToString(ciphertext), nil, 0
}

func (self *GuardServer) HeartbeatRequest(w http.ResponseWriter, req *http.Request) (interface{}, error, int) {

	body, err := utils.ReadRequestBody(req)
	if err != nil {
		log.Errorf(log.Fields{}, "HeartbeatRequest ReadRequestBody Failed [%v]", err)
		return nil, err, -1
	}

	db, err := self.newMysqlClient()
	if err != nil {
		log.Errorf(log.Fields{}, "HeartbeatRequest newMysqlClient Failed [%v]", err)
		return nil, err, -1
	}
	defer self.closeMysqlClient(db)

	mapBody := body.(map[string]interface{})
	sessionId := mapBody["sessionId"]
	softwareId := mapBody["softwareUuid"].(string)
	softwareInfo := mysql_orm.GetSoftwareDevopsStatus(db, softwareId)
	fmt.Println(sessionId, softwareId)
	fmt.Printf("%v\n", softwareInfo)
	response := make(map[string]interface{})
	response["stop"] = false
	if softwareInfo == nil || (softwareInfo != nil && softwareInfo.DevopsStatus != 0) {
		response["stop"] = true
	}
	jresponse, _ := json.Marshal(response)
	ciphertext, _ := self.rsaObjMap[sessionId.(string)].RemoteRsa.Encrypt(jresponse)
	return hex.EncodeToString(ciphertext), nil, 0
}
