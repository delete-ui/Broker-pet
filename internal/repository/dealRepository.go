package repository

import (
	"Brocker-pet-project/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
)

type DealRepository struct {
	db *sql.DB
}

func NewDealRepository(db *sql.DB) *DealRepository {
	return &DealRepository{db: db}
}

func (h *DealRepository) CreateNewDeal(title string, expenses, profit float64) *models.Deal {

	query := `INSERT INTO transactions 
    (title, expenses, profit, status) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, title, expenses, profit, status;`

	req := h.db.QueryRow(query, title, expenses, profit, "not processed")

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

func (h *DealRepository) GetAllProcessedDeals(ctx context.Context) *[]models.Deal {

	query := `SELECT * FROM transactions WHERE status=$1;`

	rows, err := h.db.QueryContext(ctx, query, "processed")
	if err != nil {
		log.Printf("Error reading sql response: %v", err)
		return nil
	}
	defer rows.Close()

	var deals []models.Deal

	for rows.Next() {
		var deal models.Deal

		if err := rows.Scan(&deal.Id, &deal.Title, &deal.Expenses, &deal.Profit, &deal.Status); err != nil {
			log.Printf("Error reading sql response: %v", err)
			return nil
		}

		deals = append(deals, deal)

	}

	if rows.Err() != nil {
		log.Printf("Error reqqading sql response: %v", rows.Err())
		return nil
	}

	return &deals
}

func (h *DealRepository) GetAllNotProcessedDeals(ctx context.Context) *[]models.Deal {

	query := `SELECT * FROM transactions WHERE status=$1;`

	rows, err := h.db.QueryContext(ctx, query, "not processed")
	if err != nil {
		log.Printf("Error reading sql response: %v", err)
		return nil
	}
	defer rows.Close()

	var deals []models.Deal

	for rows.Next() {
		var deal models.Deal

		if err := rows.Scan(&deal.Id, &deal.Title, &deal.Expenses, &deal.Profit, &deal.Status); err != nil {
			log.Printf("Error reading sql response: %v", err)
			return nil
		}

		deals = append(deals, deal)

	}

	if rows.Err() != nil {
		log.Printf("Error reqqading sql response: %v", rows.Err())
		return nil
	}

	return &deals
}

func (h *DealRepository) GetAllDeals(ctx context.Context) *[]models.Deal {
	query := `SELECT * FROM transactions WHERE id!=0;`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error reading sql response: %v", err)
		return nil
	}

	defer rows.Close()

	var deals []models.Deal

	for rows.Next() {
		var deal models.Deal

		err := rows.Scan(&deal.Id, &deal.Title, &deal.Expenses, &deal.Profit, &deal.Status)
		if err != nil {
			fmt.Printf("Error reading sql response: %v", err)
			return nil
		}

		deals = append(deals, deal)

	}

	return &deals

}

func (h *DealRepository) MarkTransactionAsProcessed(id float64) *models.Deal {

	query := `UPDATE transactions 
	SET status=$1
	WHERE id=$2
	RETURNING id, title, expenses, profit, status;`

	row := h.db.QueryRow(query, "processed", id)

	var deal models.Deal

	if err := row.Scan(&deal.Id, &deal.Title, &deal.Expenses, &deal.Profit, &deal.Status); err != nil {
		log.Printf("Error reading sql response: %v", err)
		return nil
	}

	if deal.Status != "processed" {
		log.Printf("Error marking transaction as processed: %v", deal)
		return &deal
	}

	return &deal

}
