package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/shopspring/decimal"
)

type Payslips struct {
	ID                 string          `db:"id"`
	EmployeeID         string          `db:"employee_id"`
	PayrollRunID       string          `db:"payroll_run_id"`
	AttendancePay      decimal.Decimal `db:"attendance_pay"`
	OvertimePay        decimal.Decimal `db:"overtime_pay"`
	ReimbursementTotal decimal.Decimal `db:"reimbursement_total"`
	TakeHomePay        decimal.Decimal `db:"take_home_pay"`
	GeneratedAt        time.Time       `db:"generated_at"`
	CreatedAt          time.Time       `db:"created_at"`
	UpdatedAt          time.Time       `db:"updated_at"`
	CreatedBy          sql.NullString  `db:"created_by"`
	UpdatedBy          sql.NullString  `db:"updated_by"`
	IpAddress          sql.NullString  `db:"ip_address"`
	RequestId          sql.NullString  `db:"request_id"`
}

func (p Payslips) Columns() []any {
	return []any{
		"id",
		"employee_id",
		"payroll_run_id",
		"attendance_pay",
		"overtime_pay",
		"reimbursement_total",
		"take_home_pay",
		"generated_at",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (p Payslips) StringColumns() []string {
	vals := make([]string, len(p.Columns()))

	for i, col := range p.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (p *Payslips) Values() []any {
	return []any{
		&p.ID,
		&p.EmployeeID,
		&p.PayrollRunID,
		&p.AttendancePay,
		&p.OvertimePay,
		&p.ReimbursementTotal,
		&p.TakeHomePay,
		&p.GeneratedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CreatedBy,
		&p.UpdatedBy,
		&p.IpAddress,
		&p.RequestId,
	}
}

func (p *Payslips) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(p.Values()))

	for i, val := range p.Values() {
		vals[i] = val
	}

	return vals
}
