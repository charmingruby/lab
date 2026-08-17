package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/charmingruby/new/internal/billing/model"
	"github.com/charmingruby/new/internal/billing/repository"
	"github.com/charmingruby/new/internal/billing/usecase"
	"github.com/charmingruby/new/internal/shared/core"
	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/pkg/o11y"
	"github.com/charmingruby/new/test/billing/mocks"
	sharedmocks "github.com/charmingruby/new/test/shared/mocks"
)

func TestCreatePaymentUsecase(t *testing.T) {
	o11y.InitLogger()

	tests := []struct {
		setupMocks func(
			*sharedmocks.TransactionManager[repository.Transaction],
			*mocks.OfferingRepository,
			*mocks.PaymentRepository,
			*mocks.NotificationClient,
		)
		assert func(*testing.T, usecase.CreatePaymentOutput, error)
		name   string
		input  usecase.CreatePaymentInput
	}{
		{
			name: "success",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123",
				ExternalID: "pay_123", ChargedAmount: 2999,
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offering, _ := model.NewOffering(model.OfferingInput{
					Name: "Plan", Description: "Plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999,
				})
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(offering, nil)
				txManager.On("Transact", mock.Anything).Return(
					func(fn func(repository.Transaction) error) error {
						return fn(repository.Transaction{
							OfferingRepo: offeringRepo,
							PaymentRepo:  paymentRepo,
						})
					},
				)
				paymentRepo.On("FindByExternalID", mock.Anything, "pay_123").Return(nil, nil)
				paymentRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Payment")).Return(nil)
				notifier.On("Send", mock.Anything, mock.AnythingOfType("client.SendNotificationInput")).Return(nil)
			},
			assert: func(t *testing.T, out usecase.CreatePaymentOutput, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, out.ID)
			},
		},
		{
			name: "error when offering not found",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "unknown", ExternalID: "pay_123",
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offeringRepo.On("FindByID", mock.Anything, "unknown").Return(nil, nil)
			},
			assert: func(t *testing.T, _ usecase.CreatePaymentOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsNotFound(err))
			},
		},
		{
			name: "error on offering repository failure",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123", ExternalID: "pay_123",
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(nil, errors.New("db error"))
			},
			assert: func(t *testing.T, _ usecase.CreatePaymentOutput, err error) {
				require.Error(t, err)
				assert.True(t, customerr.IsInvalidOperation(err))
			},
		},
		{
			name: "success when payment already exists (idempotent)",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123",
				ExternalID: "pay_existing", ChargedAmount: 2999,
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offering, _ := model.NewOffering(model.OfferingInput{
					Name: "Plan", Description: "Plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999,
				})
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(offering, nil)
				txManager.On("Transact", mock.Anything).Return(
					func(fn func(repository.Transaction) error) error {
						return fn(repository.Transaction{
							OfferingRepo: offeringRepo,
							PaymentRepo:  paymentRepo,
						})
					},
				)
				paymentRepo.On("FindByExternalID", mock.Anything, "pay_existing").Return(
					&model.Payment{Model: core.NewModel(), Status: model.PaidPaymentStatus}, nil,
				)
			},
			assert: func(t *testing.T, out usecase.CreatePaymentOutput, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, out.ID)
			},
		},
		{
			name: "success when notification send fails",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123",
				ExternalID: "pay_notify_fail", ChargedAmount: 2999,
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offering, _ := model.NewOffering(model.OfferingInput{
					Name: "Plan", Description: "Plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999,
				})
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(offering, nil)
				txManager.On("Transact", mock.Anything).Return(
					func(fn func(repository.Transaction) error) error {
						return fn(repository.Transaction{
							OfferingRepo: offeringRepo,
							PaymentRepo:  paymentRepo,
						})
					},
				)
				paymentRepo.On("FindByExternalID", mock.Anything, "pay_notify_fail").Return(nil, nil)
				paymentRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Payment")).Return(nil)
				notifier.On("Send", mock.Anything, mock.AnythingOfType("client.SendNotificationInput")).
					Return(errors.New("notifier down"))
			},
			assert: func(t *testing.T, out usecase.CreatePaymentOutput, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, out.ID)
			},
		},
		{
			name: "error on find by external id failure",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123",
				ExternalID: "pay_error", ChargedAmount: 2999,
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offering, _ := model.NewOffering(model.OfferingInput{
					Name: "Plan", Description: "Plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999,
				})
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(offering, nil)
				txManager.On("Transact", mock.Anything).Return(
					func(fn func(repository.Transaction) error) error {
						return fn(repository.Transaction{
							OfferingRepo: offeringRepo,
							PaymentRepo:  paymentRepo,
						})
					},
				)
				paymentRepo.On("FindByExternalID", mock.Anything, "pay_error").Return(nil, errors.New("db error"))
			},
			assert: func(t *testing.T, _ usecase.CreatePaymentOutput, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "error on create payment failure",
			input: usecase.CreatePaymentInput{
				UserID: "user-123", OfferingID: "offering-123",
				ExternalID: "pay_create_error", ChargedAmount: 2999,
			},
			setupMocks: func(txManager *sharedmocks.TransactionManager[repository.Transaction], offeringRepo *mocks.OfferingRepository, paymentRepo *mocks.PaymentRepository, notifier *mocks.NotificationClient) {
				offering, _ := model.NewOffering(model.OfferingInput{
					Name: "Plan", Description: "Plan",
					ChargeType: "one_time", Currency: "USD", Price: 2999,
				})
				offeringRepo.On("FindByID", mock.Anything, "offering-123").Return(offering, nil)
				txManager.On("Transact", mock.Anything).Return(
					func(fn func(repository.Transaction) error) error {
						return fn(repository.Transaction{
							OfferingRepo: offeringRepo,
							PaymentRepo:  paymentRepo,
						})
					},
				)
				paymentRepo.On("FindByExternalID", mock.Anything, "pay_create_error").Return(nil, nil)
				paymentRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Payment")).
					Return(errors.New("insert failed"))
			},
			assert: func(t *testing.T, _ usecase.CreatePaymentOutput, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txManager := new(sharedmocks.TransactionManager[repository.Transaction])
			offeringRepo := new(mocks.OfferingRepository)
			paymentRepo := new(mocks.PaymentRepository)
			notifier := new(mocks.NotificationClient)

			if tt.setupMocks != nil {
				tt.setupMocks(txManager, offeringRepo, paymentRepo, notifier)
			}

			uc := usecase.NewCreatePaymentUsecase(txManager, offeringRepo, paymentRepo, notifier)
			output, err := uc.CreatePayment(context.Background(), tt.input)

			if tt.assert != nil {
				tt.assert(t, output, err)
			}

			txManager.AssertExpectations(t)
			offeringRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			notifier.AssertExpectations(t)
		})
	}
}
