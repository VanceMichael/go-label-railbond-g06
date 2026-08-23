# RailBond

RailBond is a multi-tenant operations backend for bonded cross-border rail freight. It coordinates train capacity, container consignments, customs declarations, warehouse reservations, checkpoint evidence, route assignments, invoices, documents, exceptions, audit events, and retryable outbox delivery.

## Runtime

The server uses a real SQLite database, runs migrations on startup, exposes /healthz and /readyz, and accepts configuration through DATABASE_URL and PORT.

## Business flows

1. A tenant registers a corridor and train, reserves capacity, and creates a consignment with customs items.
2. Customs declarations, warehouse slots, inspections, and checkpoint evidence advance the consignment through a guarded state machine.
3. Delivery proof creates settlement work; workers deliver outbox messages and recover expired route/customs leases.
4. Every cross-entity mutation records an audit event and preserves a retryable failure state.

## Checks

Use make test, make race, make vet, and make build.

