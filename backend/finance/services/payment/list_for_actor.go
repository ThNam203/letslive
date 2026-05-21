package payment

import (
	"context"
	"sen1or/letslive/finance/dto"
	response "sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
)

func (s *PaymentService) ListForActor(ctx context.Context, actorId uuid.UUID, page int, limit int) ([]dto.PaymentResponse, int, *response.Response[any]) {
	payments, total, errResp := s.paymentRepo.ListByActor(ctx, actorId, page, limit)
	if errResp != nil {
		return nil, 0, errResp
	}

	currencies, curErr := s.currencyRepo.List(ctx)
	if curErr != nil {
		return nil, 0, curErr
	}
	precisionByCode := make(map[string]int, len(currencies))
	for _, c := range currencies {
		precisionByCode[c.Code] = c.Precision
	}

	out := make([]dto.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		out = append(out, dto.NewPaymentResponse(p, precisionByCode[p.CurrencyCode]))
	}
	return out, total, nil
}
