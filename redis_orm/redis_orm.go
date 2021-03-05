package redis_orm


import(
    log "github.com/EntropyPool/entropy-logger"
    "github.com/go-redis/redis"
)


func NewRedisClient(addr string) (*redis.Client, error){

     client := redis.NewClient(&redis.Options{
         Addr:addr,
         DB:0,
     })

     pong, err := client.Ping().Result()
     if err != nil {
        log.Errorf(log.Fields{}, "NewReidsClient error [%v]", err)
        return nil, err
     }

     if pong != "PONG" {
         log.Errorf(log.Fields{}, "Redis connect failed!")
     } else {
         log.Infof(log.Fields{}, "Redis connect success!")
     }

     return client, nil
}


