# Cross-module reads

A module never imports another module's use case. To read another module's data, the consumer depends on a **client port** owned by the producer. The producer owns three pieces per exposed read:

- **`<domain>/client/`** — the port: a read interface the consumer codes against (e.g. `PaymentReader.GetPayment`). Also holds outbound ports the module itself consumes (e.g. `NotificationClient`, adapter in `client/console`).
- **`<domain>/public/`** — the adapter: a thin struct over a use case that forwards calls into the port shape (e.g. `public.NewPaymentReader(getPayment)`, whose `GetPayment` delegates to `getPayment.GetPayment`).
- **`<domain>/public.go`** — the assembly: a module-level constructor (e.g. `NewPaymentReader(db)`) that builds repositories + use case and returns the adapter typed as the `client` port.

The consumer codes only against the produced `client` port. Mocks are generated from it into `test/<domain>/mocks`.