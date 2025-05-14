package repository

import (
	"Brocker-pet-project/internal/models"
	"database/sql"
	"log"
)

type ProfitRepository struct {
	db *sql.DB
}

func NewProfitRepository(db *sql.DB) *ProfitRepository {
	return &ProfitRepository{db}
}

func (h *ProfitRepository) AddProfitById(dealId int64, allProfit float64) *models.ProfitSQLDeal {
	query := `INSERT INTO clear_profit (deals_id,all_profit)
	VALUES ($1,$2)
	RETURNING id, deals_id,all_profit;`

	row := h.db.QueryRow(query, dealId, allProfit)

	var profit models.ProfitSQLDeal

	if err := row.Scan(&profit.Id, &profit.DealId, &profit.AllProfit); err != nil {
		log.Printf("Error parsing sql response: %v", err)
		return nil
	}

	if profit.Id == 0 {
		log.Print("Error parsing sql response")
		return nil
	}

	return &profit

}
