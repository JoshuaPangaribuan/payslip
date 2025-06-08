package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type AttendancePeriods struct {
	ID           string         `db:"id"`
	StartDate    time.Time      `db:"start_date"`
	EndDate      time.Time      `db:"end_date"`
	PayrollRun   bool           `db:"payroll_run"`
	PayrollRunID sql.NullString `db:"payroll_run_id"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	CreatedBy    sql.NullString `db:"created_by"`
	UpdatedBy    sql.NullString `db:"updated_by"`
	IpAddress    sql.NullString `db:"ip_address"`
	RequestId    sql.NullString `db:"request_id"`
}

func (ap AttendancePeriods) Columns() []any {
	return []any{
		"id",
		"start_date",
		"end_date",
		"payroll_run",
		"payroll_run_id",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (ap AttendancePeriods) StringColumns() []string {
	vals := make([]string, len(ap.Columns()))

	for i, col := range ap.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (ap *AttendancePeriods) Values() []any {
	return []any{
		&ap.ID,
		&ap.StartDate,
		&ap.EndDate,
		&ap.PayrollRun,
		&ap.PayrollRunID,
		&ap.CreatedAt,
		&ap.UpdatedAt,
		&ap.CreatedBy,
		&ap.UpdatedBy,
		&ap.IpAddress,
		&ap.RequestId,
	}
}

func (ap *AttendancePeriods) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(ap.Values()))

	for i, val := range ap.Values() {
		vals[i] = val
	}

	return vals
}
