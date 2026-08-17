# Cross-module reads

A module never imports another module's use case. To read another module's data, the consumer depends on a **client port** owned by the producer. The producer owns three pieces per exposed read:

- **`<domain>/client/`** — the port: a read interface the consumer codes against (e.g. `PaymentReader.GetPayment`). Also holds outbound ports the module itself consumes (e.g. `NotificationClient`, adapter in `client/console`). All ports for a domain — outbound and exposed reads — live together in `client/*.go`; split into separate files only when a single file grows unwieldy.
- **`<domain>/public/`** — the adapter: a thin struct over a use case that forwards calls into the port shape (e.g. `public.NewPaymentReader(getPayment)`, whose `GetPayment` delegates to `getPayment.GetPayment`).
- **`<domain>/public.go`** — the assembly: a module-level constructor (e.g. `NewPaymentReader(db)`) that builds repositories + use case and returns the adapter typed as the `client` port.

The consumer codes only against the produced `client` port. Mocks are generated from it into `test/<domain>/mocks`.

## Example — `billing` exposes `PaymentReader`

**1. The port** (`internal/billing/client/notifier.go`):

```go
type PaymentReader interface {
	GetPayment(ctx context.Context, paymentID string) (*model.Payment, error)
}
```

**2. The adapter** (`internal/billing/public/payment_reader.go`) — thin struct over a use case:

```go
type PaymentReader struct {
	getPayment usecase.GetPaymentUsecase
}

func NewPaymentReader(getPayment usecase.GetPaymentUsecase) *PaymentReader {
	return &PaymentReader{getPayment: getPayment}
}

func (r *PaymentReader) GetPayment(ctx context.Context, paymentID string) (*model.Payment, error) {
	return r.getPayment.GetPayment(ctx, usecase.GetPaymentInput{PaymentID: paymentID})
}
```

**3. The assembly** (`internal/billing/public.go`) — builds repos + use case, returns the adapter typed as the port:

```go
func NewPaymentReader(db *sqlx.DB) (client.PaymentReader, error) {
	paymentRepo, err := postgres.NewPaymentRepository(db)
	if err != nil {
		return nil, err
	}
	return public.NewPaymentReader(usecase.NewGetPaymentUsecase(paymentRepo)), nil
}
```

**4. The consumer** — depends only on `client.PaymentReader`, mocked from the port (`test/billing/mocks/PaymentReader.go`):

```go
// consumer usecase test
paymentReader := mocks.NewPaymentReader(t)
paymentReader.On("GetPayment", mock.Anything, "payment-123").
	Return(&model.Payment{Status: model.PaidPaymentStatus}, nil)
```

BAD — importing another module's internals instead of its port:

```go
// ❌ Cross-module import of the producer's internals
import "github.com/charmingruby/new/internal/billing/usecase"
import "github.com/charmingruby/new/internal/billing/model"

// ✅ Define the consumer-side interface locally, fed by billing.NewPaymentReader().
// See http route wiring that passes the billing client port into consumer modules.
```
