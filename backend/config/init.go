package config

import (
	"log"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Warning: Failed to read config file: %v", err)
	}
}
