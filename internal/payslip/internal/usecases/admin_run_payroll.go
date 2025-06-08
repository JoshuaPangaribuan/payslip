package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type (
	AdminRunPayrollUsecase interface {
		Execute(ctx context.Context, input AdminRunPayrollInput) (AdminRunPayrollOutput, error)
	}

	AdminRunPayrollInput struct {
		AttendancePeriodID uuid.UUID `json:"attendance_period_id" validate:"required"`
	}

	AdminRunPayrollOutput struct {
		PayrollRunID       uuid.UUID       `json:"payroll_run_id"`
		AttendancePeriodID uuid.UUID       `json:"attendance_period_id"`
		RunAt              time.Time       `json:"run_at"`
		TotalEmployeeCount int             `json:"total_employee_count"`
		TotalTakeHomePay   decimal.Decimal `json:"total_take_home_pay"`
	}
)
