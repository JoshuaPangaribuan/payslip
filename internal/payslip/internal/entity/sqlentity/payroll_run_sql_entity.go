package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type PayrollRuns struct {
	ID                 string         `db:"id"`
	AttendancePeriodID string         `db:"attendance_period_id"`
	RunAt              time.Time      `db:"run_at"`
	RunBy              string         `db:"run_by"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	CreatedBy          sql.NullString `db:"created_by"`
	UpdatedBy          sql.NullString `db:"updated_by"`
	IpAddress          sql.NullString `db:"ip_address"`
	RequestId          sql.NullString `db:"request_id"`
}

func (pr PayrollRuns) Columns() []any {
	return []any{
		"id",
		"attendance_period_id",
		"run_at",
		"run_by",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (pr PayrollRuns) StringColumns() []string {
	vals := make([]string, len(pr.Columns()))

	for i, col := range pr.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (pr *PayrollRuns) Values() []any {
	return []any{
		&pr.ID,
		&pr.AttendancePeriodID,
		&pr.RunAt,
		&pr.RunBy,
		&pr.CreatedAt,
		&pr.UpdatedAt,
		&pr.CreatedBy,
		&pr.UpdatedBy,
		&pr.IpAddress,
		&pr.RequestId,
	}
}

func (pr *PayrollRuns) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(pr.Values()))

	for i, val := range pr.Values() {
		vals[i] = val
	}

	return vals
}
