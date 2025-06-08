package sqlentity

import (
	"database/sql/driver"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Entity interface {
	Values() []any
	Columns() []any
	StringColumns() []string
	DriverValues() []driver.Value
	MappedValues() map[string]driver.Value
}

type UpdateEntity interface {
	MappedValues() map[string]driver.Value
}

// UUIDToString converts uuid.UUID to string
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// StringToUUID converts string to uuid.UUID
func StringToUUID(id string) uuid.UUID {
	parsed, _ := uuid.Parse(id)
	return parsed
}

// Float64ToDecimal converts float64 to decimal.Decimal
func Float64ToDecimal(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// DecimalToFloat64 converts decimal.Decimal to float64
func DecimalToFloat64(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}
