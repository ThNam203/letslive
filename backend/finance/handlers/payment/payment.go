package payment

import (
	"sen1or/letslive/finance/handlers/basehandler"
	"sen1or/letslive/finance/services/payment"
)

type PaymentHandler struct {
	basehandler.BaseHandler
	paymentService *payment.PaymentService
}

func NewPaymentHandler(paymentService *payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}
