package database

import (
	"Brocker-pet-project/internal/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode)

	var err error

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return
	}

	if err := DB.Ping(); err != nil {
		log.Printf("Error pinging database: %v", err)
	}

}

func ReturnDB() *sql.DB {
	return DB
}

func CloseDB() error {
	err := DB.Close()
	if err != nil {
		return err
	}
	return nil
}
