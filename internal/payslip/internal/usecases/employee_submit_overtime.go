package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	EmployeeSubmitOvertimeUsecase interface {
		Execute(ctx context.Context, input EmployeeSubmitOvertimeInput) (EmployeeSubmitOvertimeOutput, error)
	}

	EmployeeSubmitOvertimeInput struct {
		AttendancePeriodID uuid.UUID `json:"attendance_period_id" validate:"required"`
		Date               time.Time `json:"date" validate:"required"`
		Hours              int       `json:"hours" validate:"required,min=1,max=3"`
	}

	EmployeeSubmitOvertimeOutput struct {
		ID         uuid.UUID `json:"id"`
		EmployeeID uuid.UUID `json:"employee_id"`
		Date       time.Time `json:"date"`
		Hours      int       `json:"hours"`
		CreatedAt  time.Time `json:"created_at"`
	}
)
