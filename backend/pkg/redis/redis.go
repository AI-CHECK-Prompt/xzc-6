package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var RedisClient *redis.Client
var ctx = context.Background()

func Init() {
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}

func Close() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}

func Set(key string, value interface{}, expiration int) error {
	return RedisClient.Set(ctx, key, value, 0).Err()
}

func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func Del(key string) error {
	return RedisClient.Del(ctx, key).Err()
}
