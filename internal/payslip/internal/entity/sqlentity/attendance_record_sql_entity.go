package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type AttendanceRecords struct {
	ID                 string         `db:"id"`
	EmployeeID         string         `db:"employee_id"`
	AttendancePeriodID string         `db:"attendance_period_id"`
	AttendanceDate     time.Time      `db:"attendance_date"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	CreatedBy          sql.NullString `db:"created_by"`
	UpdatedBy          sql.NullString `db:"updated_by"`
	IpAddress          sql.NullString `db:"ip_address"`
	RequestId          sql.NullString `db:"request_id"`
}

func (ar AttendanceRecords) Columns() []any {
	return []any{
		"id",
		"employee_id",
		"attendance_period_id",
		"attendance_date",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (ar AttendanceRecords) StringColumns() []string {
	vals := make([]string, len(ar.Columns()))

	for i, col := range ar.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (ar *AttendanceRecords) Values() []any {
	return []any{
		&ar.ID,
		&ar.EmployeeID,
		&ar.AttendancePeriodID,
		&ar.AttendanceDate,
		&ar.CreatedAt,
		&ar.UpdatedAt,
		&ar.CreatedBy,
		&ar.UpdatedBy,
		&ar.IpAddress,
		&ar.RequestId,
	}
}

func (ar *AttendanceRecords) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(ar.Values()))

	for i, val := range ar.Values() {
		vals[i] = val
	}

	return vals
}
