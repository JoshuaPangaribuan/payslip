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
type EmployeeSubmitAttendanceUsecase interface {
	Execute(ctx context.Context, in usecases.EmployeeSubmitAttendanceInput) (usecases.EmployeeSubmitAttendanceOutput, error)
}

// Repository Interface
type AttendanceRepository interface {
	// CreateAttendance creates a new attendance record
	CreateAttendance(ctx context.Context, attendance *sqlentity.AttendanceRecords) error
	// GetAttendanceByEmployeeAndDate gets attendance record by employee ID and date
	GetAttendanceByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*sqlentity.AttendanceRecords, error)
	// GetAttendancePeriodByID gets attendance period by ID
	GetAttendancePeriodByID(ctx context.Context, id uuid.UUID) (*sqlentity.AttendancePeriods, error)
}

// Usecase Implementation
type employeeSubmitAttendanceInteractor struct {
	attendanceRepo AttendanceRepository
	auditLogRepo   AuditLogRepository
	logger         *zap.SugaredLogger
	uuidGen        pkguid.UUID
}

func NewEmployeeSubmitAttendanceInteractor(
	attendanceRepo AttendanceRepository,
	auditLogRepo AuditLogRepository,
	uuidGen pkguid.UUID,
	logger *zap.SugaredLogger,
) EmployeeSubmitAttendanceUsecase {
	return &employeeSubmitAttendanceInteractor{
		attendanceRepo: attendanceRepo,
		auditLogRepo:   auditLogRepo,
		uuidGen:        uuidGen,
		logger:         logger,
	}
}

func (i *employeeSubmitAttendanceInteractor) Execute(ctx context.Context, in usecases.EmployeeSubmitAttendanceInput) (usecases.EmployeeSubmitAttendanceOutput, error) {
	// Get employee ID from JWT token
	employeeIDStr := middleware.GetUserID(ctx)
	if employeeIDStr == "" {
		i.logger.Errorw("unauthorized access attempt - no user ID in token")
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("unauthorized")
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		i.logger.Errorw("invalid employee ID in token",
			"error", err,
			"employee_id_str", employeeIDStr,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("invalid employee ID")
	}

	i.logger.Infow("starting employee attendance submission",
		"employee_id", employeeID,
		"attendance_period_id", in.AttendancePeriodID,
		"date", in.Date,
	)

	// Check if date is weekend
	if in.Date.Weekday() == time.Saturday || in.Date.Weekday() == time.Sunday {
		i.logger.Warnw("attempted to submit attendance on weekend",
			"employee_id", employeeID,
			"date", in.Date,
			"weekday", in.Date.Weekday(),
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("cannot submit attendance on weekends")
	}

	// Get attendance period
	period, err := i.attendanceRepo.GetAttendancePeriodByID(ctx, in.AttendancePeriodID)
	if err != nil {
		i.logger.Errorw("failed to get attendance period",
			"error", err,
			"attendance_period_id", in.AttendancePeriodID,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewServerError("failed to get attendance period")
	}
	if period == nil {
		i.logger.Warnw("attendance period not found",
			"attendance_period_id", in.AttendancePeriodID,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("attendance period not found")
	}

	i.logger.Infow("retrieved attendance period",
		"period_id", period.ID,
		"start_date", period.StartDate,
		"end_date", period.EndDate,
	)

	// Check if date is within period range
	if in.Date.Before(period.StartDate) || in.Date.After(period.EndDate) {
		i.logger.Warnw("date outside attendance period range",
			"employee_id", employeeID,
			"date", in.Date,
			"period_start", period.StartDate,
			"period_end", period.EndDate,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("date must be within attendance period range")
	}

	// Check if attendance already exists
	existing, err := i.attendanceRepo.GetAttendanceByEmployeeAndDate(ctx, employeeID, in.Date)
	if err != nil {
		i.logger.Errorw("failed to check existing attendance",
			"error", err,
			"employee_id", employeeID,
			"date", in.Date,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewServerError("failed to check existing attendance")
	}
	if existing != nil {
		i.logger.Warnw("attendance already exists",
			"employee_id", employeeID,
			"date", in.Date,
			"existing_id", existing.ID,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewBusinessError("attendance already exists for this date")
	}

	// Create attendance record
	attendance := &sqlentity.AttendanceRecords{
		ID:                 i.uuidGen.Generate(),
		EmployeeID:         sqlentity.UUIDToString(employeeID),
		AttendancePeriodID: sqlentity.UUIDToString(in.AttendancePeriodID),
		AttendanceDate:     in.Date,
		CreatedAt:          time.Now(),
		CreatedBy:          sql.NullString{String: sqlentity.UUIDToString(employeeID), Valid: true},
	}

	i.logger.Infow("creating new attendance record",
		"attendance_id", attendance.ID,
		"employee_id", attendance.EmployeeID,
		"attendance_period_id", attendance.AttendancePeriodID,
		"date", attendance.AttendanceDate,
	)

	// Save to database
	err = i.attendanceRepo.CreateAttendance(ctx, attendance)
	if err != nil {
		i.logger.Errorw("failed to save attendance",
			"error", err,
			"attendance_id", attendance.ID,
			"employee_id", attendance.EmployeeID,
		)
		return usecases.EmployeeSubmitAttendanceOutput{}, pkgerror.NewServerError("failed to save attendance")
	}

	// Create audit log
	auditLog := &sqlentity.AuditLogs{
		ID:        i.uuidGen.Generate(),
		TableName: "attendance_record",
		RecordID:  attendance.ID,
		Action:    sqlentity.Insert,
		Changes:   sql.NullString{String: fmt.Sprintf("Created new attendance record for %s", attendance.AttendanceDate.Format("2006-01-02")), Valid: true},
		CreatedAt: time.Now(),
		CreatedBy: attendance.CreatedBy,
		IpAddress: sql.NullString{String: middleware.GetIPAddress(ctx), Valid: true},
		RequestId: sql.NullString{String: middleware.GetRequestID(ctx), Valid: true},
	}

	i.logger.Infow("creating audit log",
		"audit_log_id", auditLog.ID,
		"table_name", auditLog.TableName,
		"record_id", auditLog.RecordID,
		"action", auditLog.Action,
	)

	err = i.auditLogRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		i.logger.Errorw("failed to create audit log",
			"error", err,
			"audit_log_id", auditLog.ID,
			"record_id", auditLog.RecordID,
		)
		// Don't return error here to avoid affecting the main operation
	}

	i.logger.Infow("successfully submitted attendance",
		"attendance_id", attendance.ID,
		"employee_id", attendance.EmployeeID,
		"date", attendance.AttendanceDate,
	)

	return usecases.EmployeeSubmitAttendanceOutput{
		ID:         sqlentity.StringToUUID(attendance.ID),
		EmployeeID: sqlentity.StringToUUID(attendance.EmployeeID),
		Date:       attendance.AttendanceDate,
		CreatedAt:  attendance.CreatedAt,
	}, nil
}
