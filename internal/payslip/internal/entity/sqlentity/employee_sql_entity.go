package sqlentity

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

func (er *EmployeeRoles) Scan(value any) error {
	b, ok := value.([]byte)
	if ok {
		val := er.getMap()[string(b)]
		*er = val
		return nil
	}
	return errors.New("failed to scan employee role")
}

type Employees struct {
	ID             string          `db:"id"`
	Username       string          `db:"username"`
	PasswordHashed string          `db:"password_hash"`
	SalaryMonthly  decimal.Decimal `db:"salary_monthly"`
	Role           EmployeeRoles   `db:"role"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
	CreatedBy      sql.NullString  `db:"created_by"`
	UpdatedBy      sql.NullString  `db:"updated_by"`
	IpAddress      sql.NullString  `db:"ip_address"`
	RequestId      sql.NullString  `db:"request_id"`
}

func (e Employees) Columns() []any {
	return []any{
		"id",
		"username",
		"password_hash",
		"salary_monthly",
		"role",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"ip_address",
		"request_id",
	}
}

func (e Employees) StringColumns() []string {
	vals := make([]string, len(e.Columns()))

	for i, col := range e.Columns() {
		c, ok := col.(string)
		if !ok {
			continue
		}

		vals[i] = c
	}

	return vals
}

func (e *Employees) Values() []any {
	return []any{
		&e.ID,
		&e.Username,
		&e.PasswordHashed,
		&e.SalaryMonthly,
		&e.Role,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.CreatedBy,
		&e.UpdatedBy,
		&e.IpAddress,
		&e.RequestId,
	}
}

func (e *Employees) DriverValues() []driver.Value {
	vals := make([]driver.Value, len(e.Values()))

	for i, val := range e.Values() {
		vals[i] = val
	}

	return vals
}
