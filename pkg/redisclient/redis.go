package redis

import (
	"context"
	"github.com/jaytnw/bms-service/internal/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(cfg config.RedisConfig) {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}

func Ping() error {
	return Client.Ping(context.Background()).Err()
}
