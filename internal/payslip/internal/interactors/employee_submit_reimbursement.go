package interactors

import (
	"context"
	"database/sql"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/middleware"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkguid"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Usecase Interface
type EmployeeSubmitReimbursementUsecase interface {
	Execute(ctx context.Context, in usecases.EmployeeSubmitReimbursementInput) (usecases.EmployeeSubmitReimbursementOutput, error)
}

// Repository Interface
type ReimbursementRepository interface {
	// CreateReimbursement creates a new reimbursement request
	CreateReimbursement(ctx context.Context, reimbursement *sqlentity.ReimbursementRequests) error
}

// Usecase Implementation
type employeeSubmitReimbursementInteractor struct {
	reimbursementRepo ReimbursementRepository
	auditLogRepo      AuditLogRepository
	logger            *zap.SugaredLogger
	uuidGen           pkguid.UUID
}

func NewEmployeeSubmitReimbursementInteractor(
	reimbursementRepo ReimbursementRepository,
	auditLogRepo AuditLogRepository,
	logger *zap.SugaredLogger,
	uuidGen pkguid.UUID,
) EmployeeSubmitReimbursementUsecase {
	return &employeeSubmitReimbursementInteractor{
		reimbursementRepo: reimbursementRepo,
		auditLogRepo:      auditLogRepo,
		logger:            logger,
		uuidGen:           uuidGen,
	}
}

func (i *employeeSubmitReimbursementInteractor) Execute(ctx context.Context, in usecases.EmployeeSubmitReimbursementInput) (usecases.EmployeeSubmitReimbursementOutput, error) {
	// Get employee ID from JWT token
	employeeIDStr := middleware.GetUserID(ctx)
	if employeeIDStr == "" {
		i.logger.Errorw("unauthorized access attempt - no user ID in token")
		return usecases.EmployeeSubmitReimbursementOutput{}, pkgerror.NewBusinessError("unauthorized")
	}

	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		i.logger.Errorw("invalid employee ID in token",
			"error", err,
			"employee_id_str", employeeIDStr,
		)
		return usecases.EmployeeSubmitReimbursementOutput{}, pkgerror.NewBusinessError("invalid employee ID")
	}

	// Validate AttendancePeriodID
	if in.AttendancePeriodID == uuid.Nil {
		i.logger.Warnw("missing attendance period ID",
			"employee_id", employeeID,
		)
		return usecases.EmployeeSubmitReimbursementOutput{}, pkgerror.NewBusinessError("attendance period ID is required")
	}

	i.logger.Infow("starting reimbursement submission",
		"employee_id", employeeID,
		"attendance_period_id", in.AttendancePeriodID,
		"amount", in.Amount,
		"description", in.Description,
	)

	if in.Amount.Cmp(decimal.Zero) <= 0 {
		i.logger.Warnw("invalid reimbursement amount",
			"employee_id", employeeID,
			"amount", in.Amount,
		)
		return usecases.EmployeeSubmitReimbursementOutput{}, pkgerror.NewBusinessError("amount must be greater than 0")
	}

	reimbursement := &sqlentity.ReimbursementRequests{
		ID:                 i.uuidGen.Generate(),
		EmployeeID:         sqlentity.UUIDToString(employeeID),
		AttendancePeriodID: sqlentity.UUIDToString(in.AttendancePeriodID),
		Amount:             in.Amount,
		Description:        sql.NullString{String: in.Description, Valid: true},
		CreatedAt:          time.Now(),
		CreatedBy:          sql.NullString{String: sqlentity.UUIDToString(employeeID), Valid: true},
	}

	i.logger.Infow("creating reimbursement record",
		"reimbursement_id", reimbursement.ID,
		"employee_id", reimbursement.EmployeeID,
		"attendance_period_id", reimbursement.AttendancePeriodID,
		"amount", reimbursement.Amount,
		"description", reimbursement.Description.String,
	)

	err = i.reimbursementRepo.CreateReimbursement(ctx, reimbursement)
	if err != nil {
		i.logger.Errorw("failed to save reimbursement",
			"error", err,
			"reimbursement_id", reimbursement.ID,
			"employee_id", reimbursement.EmployeeID,
			"attendance_period_id", reimbursement.AttendancePeriodID,
		)
		return usecases.EmployeeSubmitReimbursementOutput{}, pkgerror.NewServerError("failed to save reimbursement")
	}

	i.logger.Infow("successfully saved reimbursement",
		"reimbursement_id", reimbursement.ID,
		"employee_id", reimbursement.EmployeeID,
		"attendance_period_id", reimbursement.AttendancePeriodID,
	)

	auditLogID := i.uuidGen.Generate()
	if auditLogID == "" {
		i.logger.Errorw("failed to generate audit log ID")
	} else {
		auditLog := &sqlentity.AuditLogs{
			ID:        auditLogID,
			TableName: "reimbursement_request",
			RecordID:  reimbursement.ID,
			Action:    sqlentity.Insert,
			Changes:   sql.NullString{String: "Created new reimbursement request", Valid: true},
			CreatedAt: time.Now(),
			CreatedBy: reimbursement.CreatedBy,
			IpAddress: sql.NullString{String: middleware.GetIPAddress(ctx), Valid: true},
			RequestId: sql.NullString{String: middleware.GetRequestID(ctx), Valid: true},
		}

		i.logger.Infow("creating audit log entry",
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
		}

		i.logger.Infow("successfully created audit log",
			"audit_log_id", auditLog.ID,
			"record_id", auditLog.RecordID,
		)
	}

	return usecases.EmployeeSubmitReimbursementOutput{
		ID:                 sqlentity.StringToUUID(reimbursement.ID),
		EmployeeID:         sqlentity.StringToUUID(reimbursement.EmployeeID),
		AttendancePeriodID: sqlentity.StringToUUID(reimbursement.AttendancePeriodID),
		Amount:             reimbursement.Amount,
		Description:        reimbursement.Description.String,
		CreatedAt:          reimbursement.CreatedAt,
	}, nil
}
