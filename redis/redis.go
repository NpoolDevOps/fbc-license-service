package fbcredis

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	types "github.com/NpoolDevOps/fbc-license-service/types"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
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

func (cli *RedisCli) InsertKeyInfo(keyWord string, id interface{}, info interface{}, ttl time.Duration) error {
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

type DeviceInfo struct {
	Spec      string
	SessionId uuid.UUID
}

func (cli *RedisCli) QueryDevice(spec string) (*DeviceInfo, error) {
	val, err := cli.client.Get(fmt.Sprintf("%v:device:%v", redisKeyPrefix, spec)).Result()
	if err != nil {
		return nil, err
	}
	info := &DeviceInfo{}
	err = json.Unmarshal([]byte(val), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (cli *RedisCli) QueryClient(cid uuid.UUID) (*types.ClientInfo, error) {
	val, err := cli.client.Get(fmt.Sprintf("%v:client:%v", redisKeyPrefix, cid)).Result()
	if err != nil {
		return nil, err
	}
	info := &types.ClientInfo{}
	err = json.Unmarshal([]byte(val), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (cli *RedisCli) QueryClientExpire(cid uuid.UUID) (bool, error) {
	ttl, err := cli.client.TTL(fmt.Sprintf("%v:client:%v", redisKeyPrefix, cid)).Result()
	if err != nil {
		return true, err
	}
	switch ttl.Seconds() {
	case -1:
		return false, nil
	case -2:
		return true, xerrors.Errorf("key is missed")
	default:
		if ttl < 0 {
			return true, nil
		}
	}
	return false, nil
}

type SessionInfo struct {
	SessionId    uuid.UUID
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
