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
type EmployeeSubmitOvertimeUsecase interface {
	Execute(ctx context.Context, in usecases.EmployeeSubmitOvertimeInput) (usecases.EmployeeSubmitOvertimeOutput, error)
}

// Repository Interface
type OvertimeRepository interface {
	// CreateOvertime creates a new overtime request
	CreateOvertime(ctx context.Context, overtime *sqlentity.OvertimeRequests) error
	// GetOvertimeByEmployeeAndDate gets overtime request by employee ID and date
	GetOvertimeByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*sqlentity.OvertimeRequests, error)
}

// Usecase Implementation
type employeeSubmitOvertimeInteractor struct {
	overtimeRepo   OvertimeRepository
	auditLogRepo   AuditLogRepository
	uuidGen        pkguid.UUID
	logger         *zap.SugaredLogger
	attendanceRepo AttendancePeriodRepository
}

func NewEmployeeSubmitOvertimeInteractor(
	overtimeRepo OvertimeRepository,
	auditLogRepo AuditLogRepository,
	uuidGen pkguid.UUID,
	logger *zap.SugaredLogger,
	attendanceRepo AttendancePeriodRepository,
) EmployeeSubmitOvertimeUsecase {
	return &employeeSubmitOvertimeInteractor{
		overtimeRepo:   overtimeRepo,
		auditLogRepo:   auditLogRepo,
		uuidGen:        uuidGen,
		logger:         logger,
		attendanceRepo: attendanceRepo,
	}
}

func (i *employeeSubmitOvertimeInteractor) Execute(ctx context.Context, in usecases.EmployeeSubmitOvertimeInput) (usecases.EmployeeSubmitOvertimeOutput, error) {
	i.logger.Infow("starting overtime submission process",
		"date", in.Date,
		"hours", in.Hours,
	)

	// Get employee ID from JWT token
	employeeIDStr := middleware.GetUserID(ctx)
	if employeeIDStr == "" {
		i.logger.Errorw("unauthorized access attempt - no user ID in token")
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("unauthorized")
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		i.logger.Errorw("invalid employee ID in token",
			"error", err,
			"employee_id_str", employeeIDStr,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("invalid employee ID")
	}

	i.logger.Infow("retrieved employee ID from token",
		"employee_id", employeeID,
	)

	// Validate overtime hours
	if in.Hours < 1 || in.Hours > 3 {
		i.logger.Warnw("invalid overtime hours",
			"employee_id", employeeID,
			"hours", in.Hours,
			"min_allowed", 1,
			"max_allowed", 3,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("overtime hours must be between 1 and 3")
	}

	// Get attendance period
	period, err := i.attendanceRepo.GetByID(ctx, in.AttendancePeriodID)
	if err != nil {
		i.logger.Errorw("failed to get attendance period",
			"error", err,
			"attendance_period_id", in.AttendancePeriodID,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewServerError("failed to get attendance period")
	}
	if period == nil {
		i.logger.Warnw("attendance period not found",
			"attendance_period_id", in.AttendancePeriodID,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("attendance period not found")
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
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("date must be within attendance period range")
	}

	// Check if overtime already exists
	existing, err := i.overtimeRepo.GetOvertimeByEmployeeAndDate(ctx, employeeID, in.Date)
	if err != nil {
		i.logger.Errorw("failed to check existing overtime",
			"error", err,
			"employee_id", employeeID,
			"date", in.Date,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewServerError("failed to check existing overtime")
	}
	if existing != nil {
		i.logger.Warnw("overtime already exists for date",
			"employee_id", employeeID,
			"date", in.Date,
			"existing_id", existing.ID,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewBusinessError("overtime already exists for this date")
	}

	i.logger.Infow("no existing overtime found for date",
		"employee_id", employeeID,
		"date", in.Date,
	)

	// Create overtime request
	overtimeID := i.uuidGen.Generate()
	if overtimeID == "" {
		i.logger.Errorw("failed to generate overtime ID")
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewServerError("failed to generate overtime ID")
	}

	overtime := &sqlentity.OvertimeRequests{
		ID:                 overtimeID,
		EmployeeID:         employeeIDStr,
		AttendancePeriodID: in.AttendancePeriodID.String(),
		OvertimeDate:       in.Date,
		Hours:              in.Hours,
		CreatedAt:          time.Now(),
		CreatedBy:          sql.NullString{String: employeeIDStr, Valid: true},
	}

	i.logger.Infow("creating new overtime request",
		"overtime_id", overtime.ID,
		"employee_id", overtime.EmployeeID,
		"date", overtime.OvertimeDate,
		"hours", overtime.Hours,
	)

	// Save to database
	err = i.overtimeRepo.CreateOvertime(ctx, overtime)
	if err != nil {
		i.logger.Errorw("failed to save overtime",
			"error", err,
			"overtime_id", overtime.ID,
			"employee_id", overtime.EmployeeID,
		)
		return usecases.EmployeeSubmitOvertimeOutput{}, pkgerror.NewServerError("failed to save overtime")
	}

	// Create audit log
	auditLogID := i.uuidGen.Generate()
	if auditLogID == "" {
		i.logger.Errorw("failed to generate audit log ID")
		// Don't return error here to avoid affecting the main operation
	} else {
		auditLog := &sqlentity.AuditLogs{
			ID:        auditLogID,
			TableName: "overtime_request",
			RecordID:  overtime.ID,
			Action:    sqlentity.Insert,
			Changes:   sql.NullString{String: fmt.Sprintf("Created new overtime request for %s with %d hours", overtime.OvertimeDate.Format("2006-01-02"), overtime.Hours), Valid: true},
			CreatedAt: time.Now(),
			CreatedBy: overtime.CreatedBy,
			IpAddress: sql.NullString{String: middleware.GetIPAddress(ctx), Valid: true},
			RequestId: sql.NullString{String: middleware.GetRequestID(ctx), Valid: true},
		}

		i.logger.Infow("creating audit log entry",
			"audit_log_id", auditLog.ID,
			"table_name", auditLog.TableName,
			"record_id", auditLog.RecordID,
			"action", auditLog.Action,
			"changes", auditLog.Changes.String,
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

		i.logger.Infow("successfully created audit log",
			"audit_log_id", auditLog.ID,
			"record_id", auditLog.RecordID,
		)
	}

	i.logger.Infow("overtime submission completed successfully",
		"overtime_id", overtime.ID,
		"employee_id", overtime.EmployeeID,
		"date", overtime.OvertimeDate,
		"hours", overtime.Hours,
	)

	return usecases.EmployeeSubmitOvertimeOutput{
		ID:         uuid.MustParse(overtime.ID),
		EmployeeID: uuid.MustParse(overtime.EmployeeID),
		Date:       overtime.OvertimeDate,
		Hours:      overtime.Hours,
		CreatedAt:  overtime.CreatedAt,
	}, nil
}
