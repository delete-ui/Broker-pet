package repository

import (
	"Brocker-pet-project/internal/models"
	"database/sql"
	"log"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (h *UserRepository) NewUser(username, password string) *models.NewUserResponse {
	query := `INSERT INTO users 
    (username,password) 
	VALUES ($1,$2)
	RETURNING id,username;`

	row := h.db.QueryRow(query, username, password)

	var user models.NewUserResponse

	if err := row.Scan(&user.Id, &user.Username); err != nil {
		log.Printf("Error scaning sql response: %v", err)
		return nil
	}

	if user.Id == 0 {
		log.Print("Error creating new user")
		return nil
	}

	return &user

}

func (h *UserRepository) GetUserByUsername(username, password string) *models.User {
	query := `SELECT * FROM users WHERE username=$1 AND password=$2;`

	row := h.db.QueryRow(query, username, password)

	var user models.User

	if err := row.Scan(&user.Id, &user.Username, &user.Password); err != nil {
		log.Printf("Error scanning sql response: %v", err)
		return nil
	}

	if user.Id == 0 {
		log.Printf("Error returning user by username")
		return nil
	}

	return &user

}
