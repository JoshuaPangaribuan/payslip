package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type AuditLogs struct {
	ID        string         `db:"id"`
	TableName string         `db:"table_name"`
	RecordID  string         `db:"record_id"`
	Action    AuditActions   `db:"action"`
	Changes   sql.NullString `db:"changes"` // JSONB in PostgreSQL
	CreatedAt time.Time      `db:"created_at"`
	CreatedBy sql.NullString `db:"created_by"`
	IpAddress sql.NullString `db:"ip_address"`
	RequestId sql.NullString `db:"request_id"`
}

func (al AuditLogs) Columns() []any {
	return []any{
		"id",
		"table_name",
		"record_id",
		"action",
		"changes",
		"created_at",
		"created_by",
		"ip_address",
		"request_id",
	}
}

func (al AuditLogs) StringColumns() []string {
	vals := make([]string, len(al.Columns()))

	for i, col := range al.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (al *AuditLogs) Values() []any {
	return []any{
		&al.ID,
		&al.TableName,
		&al.RecordID,
		&al.Action,
		&al.Changes,
		&al.CreatedAt,
		&al.CreatedBy,
		&al.IpAddress,
		&al.RequestId,
	}
}

func (al *AuditLogs) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(al.Values()))

	for i, val := range al.Values() {
		vals[i] = val
	}

	return vals
}
