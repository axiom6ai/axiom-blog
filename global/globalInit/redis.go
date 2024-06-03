package globalInit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func RedisInit() {
	addr := fmt.Sprintf("%s:%d", conf.Redis.Host, conf.Redis.Port)
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: conf.Redis.Password, // no password set
		DB:       0,                   // use default DB
	})
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("redis错误：%s", err))
	}
}
