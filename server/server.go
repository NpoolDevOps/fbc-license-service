package server


import (
    "fmt"
    "time"
    "net/http"
    "encoding/hex"
    "encoding/json"
    "github.com/go-basic/uuid"
    "github.com/go-redis/redis"
    "github.com/jinzhu/gorm"
    "guard_server/msg_struct"
    "guard_server/utils"
    "guard_server/redis_orm"
    "guard_server/rsa_crypto"
    "guard_server/http_daemon"
    "guard_server/mysql_orm"
)


var redisKeyPrefix = "softwareguardserver_"   // prefix + software_info_id


type GuardServer struct{
    rsaObjMap    map[string] *msg_struct.RsaPair
    authText     string
    redisAddr    string
    dbType       string
    dbUrl        string
    redisTtl     int
    redisClient *redis.Client
    mysqlClient *gorm.DB
    softwareInfoMap  map[string]*redis_orm.RedisSoftwareInfo
}


func NewGuardServer(config map[string]interface{}) *GuardServer{

    rsaObjMap := make(map[string]*msg_struct.RsaPair)
    softwareInfomap := make(map[string]*redis_orm.RedisSoftwareInfo)
    fmt.Printf("%#v\n",config)
    return &GuardServer{
        rsaObjMap: rsaObjMap,
        authText:  config["authText"].(string),
        redisAddr: config["redisAddr"].(string),
        dbType:    config["dbType"].(string),
        dbUrl:     config["dbUrl"].(string),
        redisTtl:  config["redisTtl"].(int),
        softwareInfoMap: softwareInfomap,
    }
}


func (self *GuardServer) RegisterHandler(){

    http_daemon.RegisterRouter(http_daemon.HttpRouter{
        Location:"/api_client/exchange_key",
        Handler: self.ExchangeKeyRequest,
    })

    http_daemon.RegisterRouter(http_daemon.HttpRouter{
        Location: "/api_client/startup",
        Handler: self.StartUpRequest,
    })

    http_daemon.RegisterRouter(http_daemon.HttpRouter{
        Location: "/api_client/heartbeat",
        Handler: self.HeartbeatRequest,
    })

}


func (self *GuardServer) newRedisClient(){
    if self.redisClient == nil{
        self.redisClient = redis_orm.NewRedisClient(self.redisAddr)
    }
}


func (self *GuardServer) closeRedisClient(){
    if self.redisClient != nil{
        self.redisClient.Close()
        self.redisClient = nil
    }
}


func (self *GuardServer) newMysqlClient(){

    if self.mysqlClient == nil {
        self.mysqlClient, _ = mysql_orm.NewDbOrm(self.dbType,
            self.dbUrl)
    }
}


func (self *GuardServer) closeMysqlClient(){

    if self.mysqlClient != nil{
        mysql_orm.DbOrmClose(self.mysqlClient)
        self.mysqlClient = nil
    }
}


func (self *GuardServer) Boot(){

    self.newMysqlClient()
    defer self.closeMysqlClient()

    self.newRedisClient()
    defer self.closeRedisClient()

    infos := mysql_orm.QuerySoftwareInfos(self.mysqlClient)
    for index := range infos{
        info := infos[index]
        key := redisKeyPrefix+info.Id
        redisInfo := redis_orm.RedisQuerysoftwareInfo(key, self.redisClient)
        if redisInfo != nil{
            self.softwareInfoMap[redisInfo.Id] = redisInfo
        }
    }

}


func (self *GuardServer) ExchangeKeyRequest(w http.ResponseWriter, req *http.Request)(interface{}, error, int){

    body, err := utils.ReadRequestBody(req)
    if err != nil {
        return nil, err, -1
    }

    pubkey := body.(map[string]interface{})["public_key"]
    remoteRsaObj := rsa_crypto.NewRsaCryptoWithParam([]byte(pubkey.(string)), nil)
    localRsaObj := rsa_crypto.NewRsaCrypto(1024)
    sessionId := uuid.New()
    rsaPair := &msg_struct.RsaPair{
        RemoteRsa: remoteRsaObj,
        LocalRsa: localRsaObj,
    }
    self.rsaObjMap[sessionId] = rsaPair

    response := make(map[string]interface{})
    response["sessionId"] = sessionId
    response["public_key"] = string(localRsaObj.GetPubkey())

    return response, nil, 0
}


func (self *GuardServer) StartUpRequest(w http.ResponseWriter, req *http.Request)(interface{}, error, int){

    body, err := utils.ReadRequestBody(req)
    if err != nil {
        fmt.Println(err)
        return nil, err, -1
    }

    fmt.Printf("%#v\n", body)
    sessionId := body.(map[string]interface{})["sessionId"].(string)
    data := body.(map[string]interface{})["data"].(string)
    hData,_ := hex.DecodeString(data)
    deData,_ := self.rsaObjMap[sessionId].LocalRsa.Decrypt([]byte(hData))
    var param msg_struct.StartUpReq
    err = json.Unmarshal(deData, &param)
    if err != nil {
        // log.Print(err)
    }
    fmt.Printf("param: %+v\n", param)

    self.newMysqlClient()
    defer self.closeMysqlClient()
    self.newRedisClient()
    defer self.closeRedisClient()

    response := make(map[string]interface{})
    response["authText"] = self.authText
    response["sessionId"] = sessionId
    response["startUp"] = false
    dbUserInfo := mysql_orm.QueryUserInfoBySn(self.mysqlClient, param.ClientSn)
    queryInfo := mysql_orm.QuerySoftwareInfoBySystemSn(self.mysqlClient, param.SystemSn)
    if dbUserInfo != nil{
        if queryInfo != nil {
            redisnewInfo := redis_orm.RedisSoftwareInfo{
                Id:         queryInfo.Id,
                SessionId:  sessionId,
                StartUpreq: param,
                RsaPair:    *self.rsaObjMap[sessionId],
            }
            self.softwareInfoMap[queryInfo.Id] = &redisnewInfo
            redis_orm.RedisInsertNewInfo(redisnewInfo, self.redisClient, self.redisTtl)
            response["startUp"]    = true
            response["softwareUuid"] = queryInfo.Id
        } else {
            // count := mysql_orm.GetSoftwareCount(self.mysqlClient, param.ClientSn)
            // if count < dbUserInfo.Volume{
            newInfo := mysql_orm.SoftwareInfo{
                Id: uuid.New(),
                SoftwareSn: param.ClientSn,
                SystemSn: param.SystemSn,
                Status: STATUS_UP,
                CreateTime: time.Now(),
                ModifyTime: time.Now(),
            }
            mysql_orm.InsertSoftwareInfo(self.mysqlClient, newInfo)

            redisnewInfo := redis_orm.RedisSoftwareInfo{
                Id:         newInfo.Id,
                SessionId:  sessionId,
                StartUpreq: param,
                RsaPair:    *self.rsaObjMap[sessionId],
            }
            self.softwareInfoMap[redisnewInfo.Id] = &redisnewInfo
            redis_orm.RedisInsertNewInfo(redisnewInfo, self.redisClient, self.redisTtl)
            response["startUp"]    = true
            response["softwareUuid"] = newInfo.Id
            //}
        }
    }
    fmt.Println(response)
    jresponse, _ := json.Marshal(response)
    ciphertext, _ := self.rsaObjMap[sessionId].RemoteRsa.Encrypt(jresponse)
    return hex.EncodeToString(ciphertext), nil, 0
}


func (self *GuardServer) HeartbeatRequest(w http.ResponseWriter, req *http.Request)(interface{}, error, int){

    body, err := utils.ReadRequestBody(req)
    if err != nil {
        fmt.Println(err)
        return nil, err, -1
    }

    self.newMysqlClient()
    defer self.closeMysqlClient()

    mapBody := body.(map[string]interface{})
    sessionId := mapBody["sessionId"]
    softwareId := mapBody["softwareUuid"].(string)
    softwareInfo := mysql_orm.GetSoftwareDevopsStatus(self.mysqlClient, softwareId)
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

