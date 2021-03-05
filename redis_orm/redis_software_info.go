package redis_orm


import (
    "fmt"
    "time"
    "encoding/json"
    "github.com/go-redis/redis"
    "guard_server/msg_struct"
)


var redisKeyPrefix = "softwareguardserver_"   // prefix + software_info_id


type RedisSoftwareInfo struct{
    Id            string
    SessionId     string
    RsaPair       msg_struct.RsaPair 
    StartUpreq    msg_struct.StartUpReq
}


func RedisInsertNewInfo(softwareInfo RedisSoftwareInfo, client *redis.Client, ttl int){

    jsonTask, err := json.Marshal(softwareInfo)
    if err != nil {
        fmt.Printf("json-Marshal failed: \n",err)
        return
    }
    err = client.Set(redisKeyPrefix + softwareInfo.Id, string(jsonTask), time.Duration(ttl)*time.Second).Err()
    if err != nil {
        fmt.Printf("RedisInsertNewInfo Set failed: \n", err)
        return
    }
}


func RedisQuerysoftwareInfo(key string, client *redis.Client) *RedisSoftwareInfo{

    val, err := client.Get(key).Result()
    if err != nil{
        fmt.Printf("RedisQuerysoftwareInfoFailed\n", err)
        return nil
    }
    b := []byte(val)
    info := &RedisSoftwareInfo{}
    err = json.Unmarshal(b, info)
    if err != nil {
        fmt.Printf("unmarshal failed\n", err)
        return nil
    }
    return info
}

