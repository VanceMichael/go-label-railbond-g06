package query

func checkpointTimelineScope(consignmentID string) (string, []any) {
	query := "SELECT 'checkpoint',e.evidence_hash,e.scanner_id FROM checkpoint_events e JOIN consignments c ON c.id=e.consignment_id WHERE c.id=? ORDER BY e.created_at"
	return query, []any{consignmentID}
}
