package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

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
	return RedisClient.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
}

func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func Del(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

func AcquireLock(key string, ttl int) (string, error) {
	token := generateToken()
	result, err := RedisClient.SetNX(ctx, key, token, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return "", err
	}
	if !result {
		return "", nil
	}
	return token, nil
}

func ReleaseLock(key string, token string) error {
	currentToken, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	if currentToken == token {
		return RedisClient.Del(ctx, key).Err()
	}
	return nil
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
