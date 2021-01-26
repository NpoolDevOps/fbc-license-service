package main


import (
    "fmt"
    "flag"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "encoding/hex"
    "guard_server/server"
    "guard_server/rsa_crypto"
    "guard_server/http_daemon"
)

var rsaObj = rsa_crypto.NewRsaCrypto(1024)

var authText = flag.String("authText", "The copyright belongs to npool cop.", "Auth text used at client")
var redisAddr = flag.String("redisAddr", "192.168.50.165:6379", "redis server address")
var dbType = flag.String("dbType", "mysql", "dbType")
var dbUrl = flag.String("dbUrl", "root:123456@tcp(192.168.50.165:3306)/software_guard?charset=utf8&parseTime=True&loc=Local",
 "url of mysql")
var redisTtl = flag.Int("redisTtl", 3*3600, "second for redis ttl")

func test(w http.ResponseWriter, req *http.Request)(interface{}, error, int){

    body, _ := ioutil.ReadAll(req.Body)
    var msg map[string]interface{}
    err := json.Unmarshal(body, &msg)
    if err != nil {
        return nil, err, -1
    }
    fmt.Printf("%#v\n", msg)
    text,_ := hex.DecodeString(msg["text"].(string))
    decrypto,err1 := rsaObj.Decrypt(text)
    fmt.Println("test",string(decrypto))
    fmt.Println(err1)

    return nil, nil, 0
}


func exchangeKey(w http.ResponseWriter, req *http.Request)(interface{}, error, int){

    fmt.Println("exchangeKey")
    resp := make(map[string]string)
    resp["public_key"] = string(rsaObj.GetPubkey())
    fmt.Println("exchangeKey", resp)

    return resp, nil, 0
}


func main(){

    flag.Parse()

    config := make(map[string]interface{})
    config["authText"] = *authText
    config["redisAddr"] = *redisAddr
    config["dbType"] = *dbType
    config["dbUrl"]  = *dbUrl
    config["redisTtl"] = *redisTtl
    gServer := server.NewGuardServer(config)
    gServer.Boot()
    gServer.RegisterHandler()

    http_daemon.Run(5000)

    quit := make(chan int)
     <- quit
}

