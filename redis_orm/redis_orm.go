package redis_orm


import(
    "log"
    "github.com/go-redis/redis"
)


func NewRedisClient(addr string) *redis.Client{

     client := redis.NewClient(&redis.Options{
         Addr:addr,
         DB:0,
     })

     pong, err := client.Ping().Result()
     if err != nil {
        log.Fatal(err) 
     }

     if pong != "PONG" {
         log.Fatal("Redis connect failed!")
     } else {
         log.Println("Redis connect success!")
     }

     return client
}


