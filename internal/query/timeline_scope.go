package query

func checkpointTimelineScope(tenantID, consignmentID string) (string, []any) {
	query := "SELECT 'checkpoint',e.evidence_hash,e.scanner_id FROM checkpoint_events e JOIN consignments c ON c.id=e.consignment_id WHERE e.tenant_id=? AND c.tenant_id=? AND c.id=? ORDER BY e.created_at"
	return query, []any{tenantID, tenantID, consignmentID}
}
