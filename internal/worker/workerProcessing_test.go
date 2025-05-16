package worker

import (
	"Brocker-pet-project/internal/models"
	"Brocker-pet-project/internal/repository"
	"database/sql"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return db, mock
}

func setupMockRedis() (*redis.Client, redismock.ClientMock) {
	client, mock := redismock.NewClientMock()
	return client, mock
}

func TestDealWorker_MarkAsProcessed_Success(t *testing.T) {
	// Настройка моков
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, redisMock := setupMockRedis()

	// Создаем логгер
	logger := zap.NewNop()

	// Тестовые данные
	testDeals := []models.Deal{
		{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "not processed"},
		{Id: 2, Title: "Deal 2", Expenses: 150, Profit: 300, Status: "not processed"},
	}

	// 1. Ожидание для GetAllNotProcessedDeals
	rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
		AddRow(testDeals[0].Id, testDeals[0].Title, testDeals[0].Expenses, testDeals[0].Profit, testDeals[0].Status).
		AddRow(testDeals[1].Id, testDeals[1].Title, testDeals[1].Expenses, testDeals[1].Profit, testDeals[1].Status)

	dbMock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
		WithArgs("not processed").
		WillReturnRows(rows)

	// 2. Для каждой сделки ожидаем:
	//    - сначала AddProfitById
	//    - затем MarkTransactionAsProcessed
	for i, deal := range testDeals {
		// Ожидание для AddProfitById
		profitRow := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
			AddRow(int64(i+1), deal.Id, deal.Profit-deal.Expenses)

		dbMock.ExpectQuery(`INSERT INTO clear_profit \(deals_id,all_profit\) VALUES \(\$1,\$2\) RETURNING id, deals_id,all_profit`).
			WithArgs(deal.Id, deal.Profit-deal.Expenses).
			WillReturnRows(profitRow)

		// Ожидание для MarkTransactionAsProcessed
		dealRow := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
			AddRow(deal.Id, deal.Title, deal.Expenses, deal.Profit, "processed")

		dbMock.ExpectQuery(`UPDATE transactions SET status=\$1 WHERE id=\$2 RETURNING id, title, expenses, profit, status`).
			WithArgs("processed", deal.Id).
			WillReturnRows(dealRow)
	}

	// 3. Ожидание для Redis DEL (вызывается после всех обновлений)
	redisMock.ExpectDel("notProcessedDeals:all", "processedDeals:all", "allDeals:get").SetVal(1)

	// Создаем репозитории с моками
	dealRepo := repository.NewDealRepository(db, redisClient)
	profitRepo := repository.NewProfitRepository(db)

	// Создаем worker
	worker := NewDealWorker(logger, dealRepo, profitRepo)

	// Вызываем тестируемый метод
	worker.MarkAsProcessed()

	// Проверяем, что все ожидания выполнены
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestDealWorker_MarkAsProcessed_NoDeals(t *testing.T) {
	// Настройка моков
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, _ := setupMockRedis()

	// Создаем логгер
	logger := zap.NewNop()

	// Ожидания для GetAllNotProcessedDeals - пустой результат
	rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"})
	dbMock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
		WithArgs("not processed").
		WillReturnRows(rows)

	// Создаем репозитории с моками
	dealRepo := repository.NewDealRepository(db, redisClient)
	profitRepo := repository.NewProfitRepository(db)

	// Создаем worker
	worker := NewDealWorker(logger, dealRepo, profitRepo)

	// Вызываем тестируемый метод
	worker.MarkAsProcessed()

	// Проверяем, что все ожидания выполнены
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDealWorker_MarkAsProcessed_ErrorInProfit(t *testing.T) {
	// Настройка моков
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, _ := setupMockRedis()

	// Создаем логгер
	logger := zap.NewNop()

	// Тестовые данные
	testDeal := models.Deal{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "not processed"}

	// Ожидания для GetAllNotProcessedDeals
	rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
		AddRow(testDeal.Id, testDeal.Title, testDeal.Expenses, testDeal.Profit, testDeal.Status)

	dbMock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
		WithArgs("not processed").
		WillReturnRows(rows)

	// Ожидания для AddProfitById - возвращаем ошибку
	dbMock.ExpectQuery(`INSERT INTO clear_profit`).
		WithArgs(testDeal.Id, testDeal.Profit-testDeal.Expenses).
		WillReturnError(errors.New("database error"))

	// Создаем репозитории с моками
	dealRepo := repository.NewDealRepository(db, redisClient)
	profitRepo := repository.NewProfitRepository(db)

	// Создаем worker
	worker := NewDealWorker(logger, dealRepo, profitRepo)

	// Вызываем тестируемый метод
	worker.MarkAsProcessed()

	// Проверяем, что все ожидания выполнены
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestDealWorker_MarkAsProcessed_ErrorInMarking(t *testing.T) {
	// Настройка моков
	db, dbMock := setupMockDB(t)
	defer db.Close()

	redisClient, _ := setupMockRedis()

	// Создаем логгер
	logger := zap.NewNop()

	// Тестовые данные
	testDeal := models.Deal{Id: 1, Title: "Deal 1", Expenses: 100, Profit: 200, Status: "not processed"}

	// Ожидания для GetAllNotProcessedDeals
	rows := sqlmock.NewRows([]string{"id", "title", "expenses", "profit", "status"}).
		AddRow(testDeal.Id, testDeal.Title, testDeal.Expenses, testDeal.Profit, testDeal.Status)

	dbMock.ExpectQuery(`SELECT \* FROM transactions WHERE status=\$1`).
		WithArgs("not processed").
		WillReturnRows(rows)

	// Ожидания для AddProfitById - успех
	profitRow := sqlmock.NewRows([]string{"id", "deals_id", "all_profit"}).
		AddRow(1, testDeal.Id, testDeal.Profit-testDeal.Expenses)

	dbMock.ExpectQuery(`INSERT INTO clear_profit`).
		WithArgs(testDeal.Id, testDeal.Profit-testDeal.Expenses).
		WillReturnRows(profitRow)

	// Ожидания для MarkTransactionAsProcessed - ошибка
	dbMock.ExpectQuery(`UPDATE transactions`).
		WithArgs("processed", testDeal.Id).
		WillReturnError(errors.New("update error"))

	// Создаем репозитории с моками
	dealRepo := repository.NewDealRepository(db, redisClient)
	profitRepo := repository.NewProfitRepository(db)

	// Создаем worker
	worker := NewDealWorker(logger, dealRepo, profitRepo)

	// Вызываем тестируемый метод
	worker.MarkAsProcessed()

	// Проверяем, что все ожидания выполнены
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
