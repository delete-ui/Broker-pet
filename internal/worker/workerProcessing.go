package worker

import (
	"Brocker-pet-project/internal/repository"
	"context"
	"go.uber.org/zap"
)

type DealWorker struct {
	log              *zap.Logger
	dealRepository   *repository.DealRepository
	profitRepository *repository.ProfitRepository
}

func NewDealWorker(log *zap.Logger, dealRepository *repository.DealRepository, profitRepository *repository.ProfitRepository) *DealWorker {
	return &DealWorker{log: log, dealRepository: dealRepository, profitRepository: profitRepository}
}

func (h *DealWorker) MarkAsProcessed() {
	ctx := context.Background()

	deals := h.dealRepository.GetAllNotProcessedDeals(ctx)
	if deals == nil {
		h.log.Error("Failed to get not processed deals")
		return
	}

	for _, deal := range *deals {
		profit := h.profitRepository.AddProfitById(deal.Id, deal.Profit-deal.Expenses)
		if profit == nil {
			h.log.Error("Error while adding profit for deal", zap.Int64("deal id", deal.Id))
			continue // Продолжаем обработку других сделок, а не прерываем полностью
		}

		processedDeal := h.dealRepository.MarkTransactionAsProcessed(deal.Id)
		if processedDeal == nil {
			h.log.Error("Error while marking deal as processed", zap.Int64("deal id", deal.Id))
			continue
		}

		h.log.Debug("Successfully processed deal", zap.Int64("deal id", deal.Id))
	}

	h.log.Info("Finished processing deals batch")
}
