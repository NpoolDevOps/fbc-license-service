package fbcredis

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/go-redis/redis"
	"time"
)

type RedisConfig struct {
	Host string `json:"host"`
	Ttl  int    `json:"ttl"`
}

type RedisCli struct {
	config RedisConfig
	client *redis.Client
}

func NewRedisCli(config RedisConfig) *RedisCli {
	cli := &RedisCli{
		config: config,
	}

	client := redis.NewClient(&redis.Options{
		Addr: config.Host,
		DB:   0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Errorf(log.Fields{}, "new redis client error [%v]", err)
		return nil
	}

	if pong != "PONG" {
		log.Errorf(log.Fields{}, "redis connect failed!")
	} else {
		log.Infof(log.Fields{}, "redis connect success!")
	}

	cli.client = client

	return cli
}

var redisKeyPrefix = "fbc:license:server:"

type ClientInfo struct {
	Id        string
	SessionId string
}

func (cli *RedisCli) InsertClient(info ClientInfo, ttl int) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	err = cli.client.Set(redisKeyPrefix+info.Id, string(b), time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (cli *RedisCli) QueryClient(key string) (*ClientInfo, error) {
	val, err := cli.client.Get(key).Result()
	if err != nil {
		return nil, err
	}
	info := &ClientInfo{}
	err = json.Unmarshal([]byte(val), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
