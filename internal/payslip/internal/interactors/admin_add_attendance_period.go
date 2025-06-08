package interactors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/middleware"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkguid"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Usecase Interface
type AdminAddAttendancePeriodUsecase interface {
	Execute(ctx context.Context, in usecases.AdminAddAttendancePeriodInput) (usecases.AdminAddAttendancePeriodOutput, error)
}

// Repository Interface
type AttendancePeriodRepository interface {
	// CreateAttendancePeriod creates a new attendance period
	CreateAttendancePeriod(ctx context.Context, period *sqlentity.AttendancePeriods) error
	// GetAttendancePeriodByDateRange gets attendance period by date range
	GetAttendancePeriodByDateRange(ctx context.Context, startDate, endDate time.Time) (*sqlentity.AttendancePeriods, error)
	// GetByID gets attendance period by ID
	GetByID(ctx context.Context, id uuid.UUID) (*sqlentity.AttendancePeriods, error)
}

type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, log *sqlentity.AuditLogs) error
}

// Usecase Implementation
type adminAddAttendancePeriodInteractor struct {
	attendancePeriodRepo AttendancePeriodRepository
	auditLogRepo         AuditLogRepository
	uuidGen              pkguid.UUID
	logger               *zap.SugaredLogger
}

func NewAdminAddAttendancePeriodInteractor(
	attendancePeriodRepo AttendancePeriodRepository,
	auditLogRepo AuditLogRepository,
	uuidGen pkguid.UUID,
	logger *zap.SugaredLogger,
) AdminAddAttendancePeriodUsecase {
	return &adminAddAttendancePeriodInteractor{
		attendancePeriodRepo: attendancePeriodRepo,
		auditLogRepo:         auditLogRepo,
		uuidGen:              uuidGen,
		logger:               logger,
	}
}

func (i *adminAddAttendancePeriodInteractor) Execute(ctx context.Context, in usecases.AdminAddAttendancePeriodInput) (usecases.AdminAddAttendancePeriodOutput, error) {
	// Validate date range
	if in.EndDate.Before(in.StartDate) {
		return usecases.AdminAddAttendancePeriodOutput{}, pkgerror.NewBusinessError("end date must be after start date")
	}

	// Check if period already exists
	existing, err := i.attendancePeriodRepo.GetAttendancePeriodByDateRange(ctx, in.StartDate, in.EndDate)
	if err == nil && existing != nil {
		return usecases.AdminAddAttendancePeriodOutput{}, pkgerror.NewBusinessError("attendance period already exists for this date range")
	}

	// Create attendance period
	period := &sqlentity.AttendancePeriods{
		ID:         i.uuidGen.Generate(),
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		PayrollRun: false,
		CreatedAt:  time.Now(),
		CreatedBy:  sql.NullString{String: middleware.GetUserID(ctx), Valid: true},
	}

	// Save to database
	err = i.attendancePeriodRepo.CreateAttendancePeriod(ctx, period)
	if err != nil {
		return usecases.AdminAddAttendancePeriodOutput{}, pkgerror.NewServerError("failed to save attendance period")
	}

	// Create audit log
	auditLog := &sqlentity.AuditLogs{
		ID:        i.uuidGen.Generate(),
		TableName: "attendance_period",
		RecordID:  period.ID,
		Action:    sqlentity.Insert,
		Changes:   sql.NullString{String: fmt.Sprintf("Created new attendance period from %s to %s", period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02")), Valid: true},
		CreatedAt: time.Now(),
		CreatedBy: period.CreatedBy,
		IpAddress: sql.NullString{String: middleware.GetIPAddress(ctx), Valid: true},
		RequestId: sql.NullString{String: middleware.GetRequestID(ctx), Valid: true},
	}

	err = i.auditLogRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		i.logger.Errorw("failed to create audit log", "error", err)
		// Don't return error here to avoid affecting the main operation
	}

	return usecases.AdminAddAttendancePeriodOutput{
		ID:        sqlentity.StringToUUID(period.ID),
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
		CreatedAt: period.CreatedAt,
	}, nil
}
