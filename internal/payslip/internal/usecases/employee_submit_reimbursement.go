package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type (
	EmployeeSubmitReimbursementUsecase interface {
		Execute(ctx context.Context, input EmployeeSubmitReimbursementInput) (EmployeeSubmitReimbursementOutput, error)
	}

	EmployeeSubmitReimbursementInput struct {
		EmployeeID         uuid.UUID       `json:"employee_id" validate:"required"`
		AttendancePeriodID uuid.UUID       `json:"attendance_period_id" validate:"required"`
		Amount             decimal.Decimal `json:"amount" validate:"required,gt=0"`
		Description        string          `json:"description" validate:"required"`
	}

	EmployeeSubmitReimbursementOutput struct {
		ID                 uuid.UUID       `json:"id"`
		EmployeeID         uuid.UUID       `json:"employee_id"`
		AttendancePeriodID uuid.UUID       `json:"attendance_period_id"`
		Amount             decimal.Decimal `json:"amount"`
		Description        string          `json:"description"`
		CreatedAt          time.Time       `json:"created_at"`
	}
)
