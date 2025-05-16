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

	for _, deal := range *deals {
		profit := h.profitRepository.AddProfitById(deal.Id, deal.Profit-deal.Expenses)

		if profit.Id == 0 {
			h.log.Error("Error while processing deal, ", zap.Int64("deal id: ", deal.Id))
			return
		}

		h.dealRepository.MarkTransactionAsProcessed(deal.Id)

	}

	h.log.Debug("Deals processed successfully")

}
