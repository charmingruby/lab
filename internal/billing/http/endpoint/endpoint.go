package endpoint

import "github.com/charmingruby/new/internal/billing/usecase"

type Endpoint struct {
	createOffering usecase.CreateOfferingUsecase
	createPayment  usecase.CreatePaymentUsecase
	getPayment     usecase.GetPaymentUsecase
	listPayments   usecase.ListPaymentsUsecase
}

func New(
	createOffering usecase.CreateOfferingUsecase,
	createPayment usecase.CreatePaymentUsecase,
	getPayment usecase.GetPaymentUsecase,
	listPayments usecase.ListPaymentsUsecase,
) *Endpoint {
	return &Endpoint{
		createOffering: createOffering,
		createPayment:  createPayment,
		getPayment:     getPayment,
		listPayments:   listPayments,
	}
}
