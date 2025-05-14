package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	Env      string
	Server   Server
	Worker   Worker
	Postgres Postgres
	Redis    Redis
	Jwt      Jwt
}

type Server struct {
	Host string
	Port string
}

type Worker struct {
	ProcessedTimeOut time.Duration
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Redis struct {
	Address string
}

type Jwt struct {
	Token string
}

func MustLoad(configName string) *Config {

	viper.AddConfigPath(".")
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error unmarshaling config file: %v", err)
	}

	return &cfg

}
