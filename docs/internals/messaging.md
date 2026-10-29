# Messaging

Messaging is a delivery mechanism, not an integration. A domain speaks to a broker in exactly two directions, and each one has exactly one name:

| Direction | Name          | Does                           | Wired in                          |
| --------- | ------------- | ------------------------------ | --------------------------------- |
| Outbound  | **Publisher** | writes an event to a topic     | `main`, injected into the usecase |
| Inbound   | **Listener**  | reads a topic, calls a handler | `main`, in its own goroutine      |

Never name a package, type, or file `subscriber` or `consumer` — those are broker concepts, not layer names. The broker itself is an adapter: Kafka, RabbitMQ, SQS all sit behind the same two ports.

```go
type Publisher interface {
	Publish(ctx context.Context, topic, key string, event any) error
}

type Listener[T any] interface {
	Listen(ctx context.Context, handle func(context.Context, T) error) error
}
```

Both ports live in `internal/platform/messaging/`. Provider adapters are files: `internal/platform/messaging/publisher/<provider>.go` and `internal/platform/messaging/listener/<provider>.go`. The raw broker client lives under `internal/platform/messaging/broker/<provider>/`. Full layout: [platform.md](./platform.md#messaging).

## Event contracts

The wire contract is shared by both sides, so it lives in its own package — one package per event version, one file per event:

```
event/
  <event>_v<n>/
    <event>_v<n>.go     # topic/queue name + Message
```

```go
// event/order_created_v1/order_created_v1.go
package order_created_v1

const Topic = "orders.created.v1"

type Item struct {
	SKU      string
	Quantity int
}

type Message struct {
	OrderID string
	Items   []Item
}
```

**Rules:**

- ✅ Topic name lives with the contract as a constant. Never a string literal at the publish or subscribe call site.
- ✅ Version in the package and in the topic (`order_created_v1` ↔ `orders.created.v1`).
- ✅ Only primitives and plain structs cross the boundary: timestamps as RFC 3339 strings, enums as strings, IDs as strings.
- ✅ Breaking change = new package + new topic. Both versions coexist while consumers migrate.
- ❌ No domain model, DB row, or internal type in the contract. Publisher and listener each map to and from it.
- ❌ No listener- or publisher-specific fields in the contract — it describes the event, not who reads it.

## Publisher (outbound)

- The usecase depends on the `Publisher` port, never on the broker SDK.
- The key is the ordering key: the same aggregate ID keeps related events on the same partition.
- A publish failure is mapped like any integration failure (`Integration`/`Internal`). If publishing must not fail the usecase, treat it as a non-blocking side effect and log it.
- The concrete adapter is built in `main` and passed into the module's composition root.

```go
// usecase
if err := u.pub.Publish(ctx, order_created_v1.Topic, order.ID, order_created_v1.Message{
	OrderID: order.ID,
	Items:   items,
}); err != nil {
	return customerr.Integration(err)
}
```

## Listener (inbound)

One package per consumed event, versioned in the folder and package name:

```
<module>/
  listener/
    on_<event>_v<n>/
      listener.go       # Listener + New + Handle
      ...               # parsing, state machine, persistence for this event
```

- `Listener` is the struct, `New(deps...) (*Listener, error)` the constructor, `Handle(ctx, <event>.Message) error` the only entry point the transport calls.
- `New` validates what it can at boot so bad wiring fails startup, not the first message.
- The listener package never owns the wire type and never imports the broker SDK — the listener adapter in `internal/platform/messaging/listener/<provider>.go` decodes and calls `Handle`.

Listener only → keep it flat, no wrapper. HTTP and an inbound listener in the same module → both move under `delivery/` (`delivery/http/`, `delivery/queue/` with the listener packages inside). A publisher alone does not trigger the split: an HTTP-only module that publishes stays flat (`http/`). See [architecture.md](./architecture.md#delivery-mechanism-layout).

```go
type Listener struct {
	repo      *repository.OrderRepository
	txManager *repository.TransactionManager
	machine   *fsm.FSM[machineState]
}

func New(repo *repository.OrderRepository, txManager *repository.TransactionManager) (*Listener, error) {
	// build + validate
}

func (l *Listener) Handle(ctx context.Context, message order_created_v1.Message) error {
	// one fresh run state per message, then delegate
}
```

**Failure semantics:**

- Returning an error stops consumption and skips the ack/commit → the message is redelivered. Use it for infrastructure failures and unparseable messages.
- A failure scoped to one record is logged and skipped inside `Handle`; the message is committed and the batch continues.
- Delivery is at-least-once, so `Handle` must be idempotent (upserts, replace-by-key, dedupe).

## Wiring in main

The module's composition root exposes every listener it owns through one `Listeners` struct. Building a listener is wiring, not startup — nothing consumes yet.

```go
// <module>/<module>.go
type Listeners struct {
	OnOrderCreatedV1 *on_order_created_v1.Listener
	OnOrderPaidV1    *on_order_paid_v1.Listener
}

func New(db *sqlx.DB, pub messaging.Publisher) (*Listeners, error) {
	// repositories, transaction manager, then each listener
}
```

`Listen` blocks — it loops over the topic until the context is canceled. **Register each listener in `main`, one goroutine per topic.** Two listeners in the same goroutine means the second topic is never consumed.

```go
// cmd/<service>/main.go
listeners, err := <module>.New(db, pub)
if err != nil {
	return err
}

shutdownErrCh := make(chan error, 1)

go func() {
	if err := listener.New[order_created_v1.Message](reader(order_created_v1.Topic)).
		Listen(ctx, listeners.OnOrderCreatedV1.Handle); err != nil {
		shutdownErrCh <- err
	}
}()

go func() {
	if err := listener.New[order_paid_v1.Message](reader(order_paid_v1.Topic)).
		Listen(ctx, listeners.OnOrderPaidV1.Handle); err != nil {
		shutdownErrCh <- err
	}
}()
```

The same `ctx` from `signal.NotifyContext` stops every listener on SIGINT/SIGTERM. A failing listener pushes to the error channel so `main` exits non-zero instead of dying silently. Resource cleanup (DB, broker clients) stays in `shutdown`.

## Rules

- ✅ A subscriber is a `listener`, in `listener/on_<event>_v<n>/`, with `Listener`, `New`, `Handle`.
- ✅ Event contracts live in their own versioned package, with the topic name as a constant.
- ✅ One listener package per event version; a breaking payload change adds `on_<event>_v2`, it never mutates v1.
- ✅ Register each listener in `main`, one goroutine per topic, sharing the signal context.
- ✅ Report listener failures through an error channel so `main` exits non-zero.
- ✅ Keep `Handle` thin: decoding stays in the adapter, business work goes to the module's usecase or state machine.
- ❌ Never call `Listen` inline in `main` — it never returns and the process stops booting.
- ❌ Never start listeners inside `<module>.go` or a usecase — the composition root only builds.
- ❌ Never put a listener behind a `client/` port or a `repository` — it is inbound delivery.
- ❌ Never run two topics in one goroutine — the second one starves.
- ❌ Never leak broker types (message, offset, headers) past the adapter.

## Related

- [Architecture](./architecture.md#delivery-mechanism-layout) — where the mechanism folder lives
- [Platform](./platform.md#messaging) — ports, adapters, and the raw broker client
- [External integrations](./external-integrations.md) — why messaging is not a `client`
- [Versioning](./versioning.md) — event versions at the caller-facing boundary
