package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type (
	AdminPayrollSummaryUsecase interface {
		Execute(ctx context.Context, input AdminPayrollSummaryInput) (AdminPayrollSummaryOutput, error)
	}

	AdminPayrollSummaryInput struct {
		PayrollRunID uuid.UUID `json:"payroll_run_id" validate:"required"`
	}

	EmployeePayrollSummary struct {
		EmployeeID     uuid.UUID       `json:"employee_id"`
		EmployeeName   string          `json:"employee_name"`
		BaseSalary     decimal.Decimal `json:"base_salary"`
		OvertimePay    decimal.Decimal `json:"overtime_pay"`
		Reimbursements decimal.Decimal `json:"reimbursements"`
		TakeHomePay    decimal.Decimal `json:"take_home_pay"`
	}

	AdminPayrollSummaryOutput struct {
		PayrollRunID       uuid.UUID                `json:"payroll_run_id"`
		AttendancePeriodID uuid.UUID                `json:"attendance_period_id"`
		ProcessedAt        time.Time                `json:"processed_at"`
		EmployeeSummaries  []EmployeePayrollSummary `json:"employee_summaries"`
		TotalTakeHomePay   decimal.Decimal          `json:"total_take_home_pay"`
	}
)
