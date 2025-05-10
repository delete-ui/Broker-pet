package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Env      string
	Server   Server
	Postgres Postgres
}

type Server struct {
	Host string
	Port string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
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
