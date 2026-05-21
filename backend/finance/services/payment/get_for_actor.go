package payment

import (
	"context"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
)

func (s *PaymentService) GetForActor(ctx context.Context, paymentId uuid.UUID, actorId uuid.UUID) (*dto.PaymentResponse, *response.Response[any]) {
	p, errResp := s.paymentRepo.GetById(ctx, paymentId)
	if errResp != nil {
		return nil, errResp
	}

	tx, txErr := s.transactionRepo.GetById(ctx, p.TransactionId)
	if txErr != nil {
		return nil, txErr
	}
	if tx.ActorId == nil || *tx.ActorId != actorId {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_PAYMENT_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	currency, curErr := s.currencyRepo.GetByCode(ctx, p.CurrencyCode)
	if curErr != nil {
		return nil, curErr
	}

	resp := dto.NewPaymentResponse(*p, currency.Precision)
	return &resp, nil
}
