package sqlentity

import (
	"database/sql/driver"
	"errors"
)

type EmployeeRoles int

const (
	Employee EmployeeRoles = iota
	Admin
)

func (er EmployeeRoles) String() string {
	return []string{"employee", "admin"}[er]
}

func (er EmployeeRoles) Value() (driver.Value, error) {
	return er.String(), nil
}

func (er EmployeeRoles) getMap() map[string]EmployeeRoles {
	return map[string]EmployeeRoles{
		"employee": Employee,
		"admin":    Admin,
	}
}

type AuditActions int

const (
	Insert AuditActions = iota
	Update
	Delete
)

func (aa AuditActions) String() string {
	return []string{"INSERT", "UPDATE", "DELETE"}[aa]
}

func (aa AuditActions) Value() (driver.Value, error) {
	return aa.String(), nil
}

func (aa AuditActions) getMap() map[string]AuditActions {
	return map[string]AuditActions{
		"INSERT": Insert,
		"UPDATE": Update,
		"DELETE": Delete,
	}
}

func (aa *AuditActions) Scan(value any) error {
	b, ok := value.([]byte)
	if ok {
		val := aa.getMap()[string(b)]
		*aa = val
		return nil
	}
	return errors.New("failed to scan audit action")
}
