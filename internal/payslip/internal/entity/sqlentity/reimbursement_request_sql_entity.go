package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/shopspring/decimal"
)

type ReimbursementRequests struct {
	ID                 string          `db:"id"`
	EmployeeID         string          `db:"employee_id"`
	AttendancePeriodID string          `db:"attendance_period_id"`
	Amount             decimal.Decimal `db:"amount"`
	Description        sql.NullString  `db:"description"`
	CreatedAt          time.Time       `db:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at"`
	CreatedBy          sql.NullString  `db:"created_by"`
	UpdatedBy          sql.NullString  `db:"updated_by"`
	IpAddress          sql.NullString  `db:"ip_address"`
	RequestId          sql.NullString  `db:"request_id"`
}

func (rr ReimbursementRequests) Columns() []any {
	return []any{
		"id",
		"employee_id",
		"attendance_period_id",
		"amount",
		"description",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (rr ReimbursementRequests) StringColumns() []string {
	vals := make([]string, len(rr.Columns()))

	for i, col := range rr.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (rr *ReimbursementRequests) Values() []any {
	return []any{
		&rr.ID,
		&rr.EmployeeID,
		&rr.AttendancePeriodID,
		&rr.Amount,
		&rr.Description,
		&rr.CreatedAt,
		&rr.UpdatedAt,
		&rr.CreatedBy,
		&rr.UpdatedBy,
		&rr.IpAddress,
		&rr.RequestId,
	}
}

func (rr *ReimbursementRequests) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(rr.Values()))

	for i, val := range rr.Values() {
		vals[i] = val
	}

	return vals
}
