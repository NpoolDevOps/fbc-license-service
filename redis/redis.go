package fbcredis

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"time"
)

type RedisConfig struct {
	Host string        `json:"host"`
	Ttl  time.Duration `json:"ttl"`
}

type RedisCli struct {
	config RedisConfig
	client *redis.Client
}

func NewRedisCli(config RedisConfig) *RedisCli {
	cli := &RedisCli{
		config: config,
	}

	var myConfig RedisConfig

	resp, err := etcdcli.Get(config.Host)
	if err == nil {
		err = json.Unmarshal(resp[0], &myConfig)
		if err == nil {
			cli = &RedisCli{
				config: myConfig,
			}
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr: cli.config.Host,
		DB:   0,
	})

	log.Infof(log.Fields{}, "redis ping -> %v", config.Host)
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
	SessionId uuid.UUID
	Id        uuid.UUID
}

func (cli *RedisCli) InsertKeyInfo(keyWord string, id uuid.UUID, info interface{}, ttl time.Duration) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	err = cli.client.Set(fmt.Sprintf("%v:%v:%v", redisKeyPrefix, keyWord, id),
		string(b), ttl*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (cli *RedisCli) QueryClient(cid uuid.UUID) (*ClientInfo, error) {
	val, err := cli.client.Get(fmt.Sprintf("%v:client:%v", redisKeyPrefix, cid)).Result()
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

type SessionInfo struct {
	SessionId    string
	MyPubKey     string
	ClientPubKey string
}

func (cli *RedisCli) QuerySession(sid uuid.UUID) (*SessionInfo, error) {
	val, err := cli.client.Get(fmt.Sprintf("%v:session:%v", redisKeyPrefix, sid)).Result()
	if err != nil {
		return nil, err
	}
	info := &SessionInfo{}
	err = json.Unmarshal([]byte(val), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
