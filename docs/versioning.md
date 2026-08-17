# Versioning

- **Unit of versioning: the endpoint file, not the folder.** A breaking change to `create_payment_v1.go` adds `create_payment_v2.go` beside it. The rest of `endpoint/` stays on v1.
- **`route.go` never carries a version suffix.** It's the single registrar mapping every endpoint version to its route (`/api/v1/...`, `/api/v2/...`). Deprecate `/v1` by removing its registration once callers migrate — don't mutate `/v1`'s behavior in place.
- **Non-HTTP protocols version the same way**, at the caller-facing boundary: gRPC registers `ServiceV1`/`ServiceV2`; a queue consumer binds versioned events or queue names.
- **Never version `model/`, `usecase/`, or `repository/`.** These stay single-version regardless of how many API versions call them.