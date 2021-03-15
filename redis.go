package main

type RedisConfig struct {
	Host string `json:"host"`
	Ttl  int    `json:"ttl"`
}

type RedisCli struct {
	config RedisConfig
}

func NewRedisCli(config RedisConfig) *RedisCli {
	cli := &RedisCli{
		config: config,
	}
	return cli
}
