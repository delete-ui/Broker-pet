package repository

import (
	"Brocker-pet-project/internal/models"
	"database/sql"
	"log"
)

type DealRepository struct {
	db *sql.DB
}

func NewDealRepository(db *sql.DB) *DealRepository {
	return &DealRepository{db: db}
}

func (h *DealRepository) CreateNewDeal(title string, expenses, profit float64) *models.Deal {
	query := `INSERT INTO transaction 
    (title, expenses, profit) 
	VALUES ($1, $2, $3)
	RETURNING id, title, expenses, profit, status;`

	req := h.db.QueryRow(query, title, expenses, profit)

	var deal models.Deal

	if err := req.Scan(&deal.Id, &deal.Title, &deal.Expenses, &deal.Profit, &deal.Status); err != nil {
		log.Printf("Error scaning sql response: %v", err)
		return nil
	}

	if deal.Id == 0 {
		log.Printf("Error inserting new deal: %v", deal)
		return nil
	}

	return &deal
}
