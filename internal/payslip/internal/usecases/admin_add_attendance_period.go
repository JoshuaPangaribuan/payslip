package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	AdminAddAttendancePeriodUsecase interface {
		Execute(ctx context.Context, input AdminAddAttendancePeriodInput) (AdminAddAttendancePeriodOutput, error)
	}

	AdminAddAttendancePeriodInput struct {
		StartDate time.Time `json:"start_date" validate:"required"`
		EndDate   time.Time `json:"end_date" validate:"required"`
	}

	AdminAddAttendancePeriodOutput struct {
		ID        uuid.UUID `json:"id"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
		CreatedAt time.Time `json:"created_at"`
	}
)
