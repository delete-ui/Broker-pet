package repository

import (
	"Brocker-pet-project/internal/models"
	"context"
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

func (h *ProfitRepository) GetAllProfitInfo(ctx context.Context) *[]models.ProfitSQLDeal {
	query := `SELECT id, deals_id, all_profit FROM clear_profit;`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing sql query: %v", err)
		return nil
	}
	defer rows.Close()

	var profits []models.ProfitSQLDeal

	for rows.Next() {
		var profit models.ProfitSQLDeal
		if err := rows.Scan(&profit.Id, &profit.DealId, &profit.AllProfit); err != nil {
			log.Printf("Error scanning sql response: %v", err)
			return nil
		}
		profits = append(profits, profit)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after scanning rows: %v", err)
		return nil
	}

	return &profits
}
