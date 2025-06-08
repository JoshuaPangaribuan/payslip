package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	EmployeeSubmitAttendanceUsecase interface {
		Execute(ctx context.Context, input EmployeeSubmitAttendanceInput) (EmployeeSubmitAttendanceOutput, error)
	}

	EmployeeSubmitAttendanceInput struct {
		AttendancePeriodID uuid.UUID `json:"attendance_period_id" validate:"required"`
		Date               time.Time `json:"date" validate:"required"`
	}

	EmployeeSubmitAttendanceOutput struct {
		ID         uuid.UUID `json:"id"`
		EmployeeID uuid.UUID `json:"employee_id"`
		Date       time.Time `json:"date"`
		CreatedAt  time.Time `json:"created_at"`
	}
)
