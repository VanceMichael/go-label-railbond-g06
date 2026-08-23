package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TrainRow struct {
	ID, TenantID, CorridorID, Number, Status string
	Capacity, Reserved, Version              int
	SlotID                                   sql.NullString
	DepartureAt                              time.Time
}

type ConsignmentRow struct {
	ID, TenantID, TrainID, ContainerID, Reference, Status string
	CurrentCheckpoint, Version                            int
	DeliveredAt, ArchivedAt                               sql.NullTime
}

type DeclarationRow struct {
	ID, TenantID, ConsignmentID, Status string
	HoldReason, BrokerOperationKey      sql.NullString
	Version                             int
}

func (s *Store) GetTrain(ctx context.Context, tenantID, id string) (TrainRow, error) {
	var r TrainRow
	var departure string
	err := s.QueryRow(ctx, "SELECT id,tenant_id,corridor_id,number,status,capacity,reserved,version,slot_id,departure_at FROM trains WHERE tenant_id=? AND id=?", tenantID, id).
		Scan(&r.ID, &r.TenantID, &r.CorridorID, &r.Number, &r.Status, &r.Capacity, &r.Reserved, &r.Version, &r.SlotID, &departure)
	if err != nil {
		if IsMissing(err) {
			return r, fmt.Errorf("%w: train", sql.ErrNoRows)
		}
		return r, err
	}
	r.DepartureAt, _ = time.Parse(time.RFC3339Nano, departure)
	return r, nil
}

func (s *Store) GetConsignment(ctx context.Context, tenantID, id string) (ConsignmentRow, error) {
	var r ConsignmentRow
	var delivered, archived sql.NullString
	err := s.QueryRow(ctx, "SELECT id,tenant_id,train_id,container_id,reference,status,current_checkpoint,version,delivered_at,archived_at FROM consignments WHERE tenant_id=? AND id=?", tenantID, id).
		Scan(&r.ID, &r.TenantID, &r.TrainID, &r.ContainerID, &r.Reference, &r.Status, &r.CurrentCheckpoint, &r.Version, &delivered, &archived)
	if err != nil {
		return r, err
	}
	if delivered.Valid {
		r.DeliveredAt = sql.NullTime{Time: parseTime(delivered.String), Valid: true}
	}
	if archived.Valid {
		r.ArchivedAt = sql.NullTime{Time: parseTime(archived.String), Valid: true}
	}
	return r, nil
}

func (s *Store) GetDeclaration(ctx context.Context, tenantID, id string) (DeclarationRow, error) {
	var r DeclarationRow
	err := s.QueryRow(ctx, "SELECT id,tenant_id,consignment_id,status,hold_reason,broker_operation_key,version FROM customs_declarations WHERE tenant_id=? AND id=?", tenantID, id).
		Scan(&r.ID, &r.TenantID, &r.ConsignmentID, &r.Status, &r.HoldReason, &r.BrokerOperationKey, &r.Version)
	return r, err
}

func parseTime(value string) time.Time { t, _ := time.Parse(time.RFC3339Nano, value); return t }

func (s *Store) Count(ctx context.Context, table, tenantID string) (int, error) {
	var n int
	err := s.QueryRow(ctx, "SELECT COUNT(*) FROM "+table+" WHERE tenant_id=?", tenantID).Scan(&n)
	return n, err
}
