package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

type OvertimeRequests struct {
	ID                 string         `db:"id"`
	EmployeeID         string         `db:"employee_id"`
	AttendancePeriodID string         `db:"attendance_period_id"`
	OvertimeDate       time.Time      `db:"overtime_date"`
	Hours              int            `db:"hours"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	CreatedBy          sql.NullString `db:"created_by"`
	UpdatedBy          sql.NullString `db:"updated_by"`
	IpAddress          sql.NullString `db:"ip_address"`
	RequestId          sql.NullString `db:"request_id"`
}

func (or OvertimeRequests) Columns() []any {
	return []any{
		"id",
		"employee_id",
		"attendance_period_id",
		"overtime_date",
		"hours",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (or OvertimeRequests) StringColumns() []string {
	vals := make([]string, len(or.Columns()))

	for i, col := range or.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (or *OvertimeRequests) Values() []any {
	return []any{
		&or.ID,
		&or.EmployeeID,
		&or.AttendancePeriodID,
		&or.OvertimeDate,
		&or.Hours,
		&or.CreatedAt,
		&or.UpdatedAt,
		&or.CreatedBy,
		&or.UpdatedBy,
		&or.IpAddress,
		&or.RequestId,
	}
}

func (or *OvertimeRequests) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(or.Values()))

	for i, val := range or.Values() {
		vals[i] = val
	}

	return vals
}
