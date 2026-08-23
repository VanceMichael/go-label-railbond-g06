PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tenants (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, timezone TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), email TEXT NOT NULL,
 role TEXT NOT NULL, password_hash TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL,
 UNIQUE(tenant_id,email)
);
CREATE TABLE IF NOT EXISTS sessions (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id), token_hash TEXT NOT NULL UNIQUE,
 expires_at TEXT NOT NULL, revoked_at TEXT, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS corridors (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), name TEXT NOT NULL,
 origin TEXT NOT NULL, destination TEXT NOT NULL, timezone TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 1,
 created_at TEXT NOT NULL, UNIQUE(tenant_id,name)
);
CREATE TABLE IF NOT EXISTS trains (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), corridor_id TEXT NOT NULL REFERENCES corridors(id),
 number TEXT NOT NULL, status TEXT NOT NULL, capacity INTEGER NOT NULL, reserved INTEGER NOT NULL DEFAULT 0,
 version INTEGER NOT NULL DEFAULT 1, slot_id TEXT, departure_at TEXT NOT NULL, created_at TEXT NOT NULL,
 UNIQUE(tenant_id,number)
);
CREATE TABLE IF NOT EXISTS rail_slots (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), corridor_id TEXT NOT NULL REFERENCES corridors(id),
 starts_at TEXT NOT NULL, ends_at TEXT NOT NULL, status TEXT NOT NULL, train_id TEXT REFERENCES trains(id),
 version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS containers (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), code TEXT NOT NULL,
 status TEXT NOT NULL, lease_owner TEXT, lease_token TEXT, lease_until TEXT, version INTEGER NOT NULL DEFAULT 1,
 created_at TEXT NOT NULL, UNIQUE(tenant_id,code)
);
CREATE TABLE IF NOT EXISTS consignments (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), train_id TEXT NOT NULL REFERENCES trains(id),
 container_id TEXT NOT NULL REFERENCES containers(id), reference TEXT NOT NULL, status TEXT NOT NULL,
 current_checkpoint INTEGER NOT NULL DEFAULT 0, version INTEGER NOT NULL DEFAULT 1,
 created_at TEXT NOT NULL, delivered_at TEXT, archived_at TEXT,
 UNIQUE(tenant_id,reference)
);
CREATE TABLE IF NOT EXISTS consignment_items (
 id TEXT PRIMARY KEY, consignment_id TEXT NOT NULL REFERENCES consignments(id), sku TEXT NOT NULL,
 description TEXT NOT NULL, quantity INTEGER NOT NULL, declared_value INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS customs_declarations (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), consignment_id TEXT NOT NULL REFERENCES consignments(id),
 status TEXT NOT NULL, hold_reason TEXT, broker_operation_key TEXT, submitted_at TEXT, released_at TEXT,
 version INTEGER NOT NULL DEFAULT 1, UNIQUE(consignment_id)
);
CREATE TABLE IF NOT EXISTS warehouse_slots (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), code TEXT NOT NULL,
 zone TEXT NOT NULL, status TEXT NOT NULL, reserved_for TEXT, version INTEGER NOT NULL DEFAULT 1,
 UNIQUE(tenant_id,code)
);
CREATE TABLE IF NOT EXISTS slot_reservations (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), slot_id TEXT NOT NULL REFERENCES warehouse_slots(id),
 consignment_id TEXT NOT NULL REFERENCES consignments(id), status TEXT NOT NULL, created_at TEXT NOT NULL,
 released_at TEXT, UNIQUE(slot_id,consignment_id)
);
CREATE TABLE IF NOT EXISTS checkpoints (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), corridor_id TEXT NOT NULL REFERENCES corridors(id),
 sequence_no INTEGER NOT NULL, name TEXT NOT NULL, required_inspection INTEGER NOT NULL DEFAULT 0,
 UNIQUE(corridor_id,sequence_no)
);
CREATE TABLE IF NOT EXISTS checkpoint_events (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), checkpoint_id TEXT NOT NULL REFERENCES checkpoints(id),
 consignment_id TEXT NOT NULL REFERENCES consignments(id), scanner_id TEXT NOT NULL, evidence_hash TEXT NOT NULL,
 observed_at TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(consignment_id,checkpoint_id)
);
CREATE TABLE IF NOT EXISTS inspections (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), checkpoint_id TEXT NOT NULL REFERENCES checkpoints(id),
 consignment_id TEXT NOT NULL REFERENCES consignments(id), status TEXT NOT NULL, result TEXT, completed_at TEXT,
 UNIQUE(consignment_id,checkpoint_id)
);
CREATE TABLE IF NOT EXISTS invoices (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), consignment_id TEXT NOT NULL REFERENCES consignments(id),
 status TEXT NOT NULL, amount INTEGER NOT NULL, dispute_reason TEXT, issued_at TEXT NOT NULL,
 settled_at TEXT, UNIQUE(consignment_id)
);
CREATE TABLE IF NOT EXISTS payments (
 id TEXT PRIMARY KEY, invoice_id TEXT NOT NULL REFERENCES invoices(id), provider_key TEXT NOT NULL,
 status TEXT NOT NULL, amount INTEGER NOT NULL DEFAULT 0, captured_at TEXT, UNIQUE(provider_key)
);
CREATE TABLE IF NOT EXISTS exceptions (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), consignment_id TEXT NOT NULL REFERENCES consignments(id),
 status TEXT NOT NULL, reason TEXT NOT NULL, replacement_route TEXT, opened_at TEXT NOT NULL, resolved_at TEXT
);
CREATE TABLE IF NOT EXISTS route_assignments (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), consignment_id TEXT NOT NULL REFERENCES consignments(id),
 carrier TEXT NOT NULL, status TEXT NOT NULL, lease_owner TEXT, lease_epoch INTEGER NOT NULL DEFAULT 1,
 lease_until TEXT, attempt INTEGER NOT NULL DEFAULT 0, next_attempt_at TEXT NOT NULL, last_error TEXT
);
CREATE TABLE IF NOT EXISTS documents (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), consignment_id TEXT NOT NULL REFERENCES consignments(id),
 kind TEXT NOT NULL, status TEXT NOT NULL, content_hash TEXT, version INTEGER NOT NULL DEFAULT 1,
 created_at TEXT NOT NULL, sealed_at TEXT
);
CREATE TABLE IF NOT EXISTS outbox_messages (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), topic TEXT NOT NULL,
 aggregate_id TEXT NOT NULL, payload TEXT NOT NULL, status TEXT NOT NULL, lease_owner TEXT,
 lease_epoch INTEGER NOT NULL DEFAULT 1, lease_until TEXT, attempts INTEGER NOT NULL DEFAULT 0,
 available_at TEXT NOT NULL, last_error TEXT, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS audit_events (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), actor_id TEXT NOT NULL,
 action TEXT NOT NULL, object_type TEXT NOT NULL, object_id TEXT NOT NULL, outcome TEXT NOT NULL,
 request_id TEXT NOT NULL, detail TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS idempotency_keys (
 id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL REFERENCES tenants(id), key TEXT NOT NULL,
 method TEXT NOT NULL, path TEXT NOT NULL, status_code INTEGER NOT NULL, response_body TEXT NOT NULL,
 created_at TEXT NOT NULL, UNIQUE(tenant_id,key,method,path)
);
CREATE TABLE IF NOT EXISTS worker_leases (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, owner TEXT NOT NULL, epoch INTEGER NOT NULL DEFAULT 1,
 lease_until TEXT NOT NULL, state TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(name)
);
CREATE INDEX IF NOT EXISTS idx_consignments_tenant_status ON consignments(tenant_id,status);
CREATE INDEX IF NOT EXISTS idx_outbox_ready ON outbox_messages(status,available_at);
CREATE INDEX IF NOT EXISTS idx_assignments_ready ON route_assignments(status,next_attempt_at);
CREATE INDEX IF NOT EXISTS idx_audit_object ON audit_events(tenant_id,object_type,object_id,created_at);
