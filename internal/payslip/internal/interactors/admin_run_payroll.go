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
)

// Usecase Interface
type AdminRunPayrollUsecase interface {
	Execute(ctx context.Context, in usecases.AdminRunPayrollInput) (usecases.AdminRunPayrollOutput, error)
}

// Repository Interface
type PayrollRepository interface {
	// CreatePayrollRun creates a new payroll run
	CreatePayrollRun(ctx context.Context, payrollRun *sqlentity.PayrollRuns) error
	// GetPayrollRunByAttendancePeriod gets payroll run by attendance period ID
	GetPayrollRunByAttendancePeriod(ctx context.Context, attendancePeriodID uuid.UUID) (*sqlentity.PayrollRuns, error)
	// GetActiveEmployees gets all active employees
	GetActiveEmployees(ctx context.Context) ([]*sqlentity.Employees, error)
	// GetAttendanceByPeriod gets attendance records for a period
	GetAttendanceByPeriod(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.AttendanceRecords, error)
	// GetOvertimeByPeriod gets overtime requests for a period
	GetOvertimeByPeriod(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.OvertimeRequests, error)
	// GetReimbursementsByPeriod gets reimbursement requests for a period
	GetReimbursementsByPeriod(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.ReimbursementRequests, error)
}

// Usecase Implementation
type adminRunPayrollInteractor struct {
	payrollRepo          PayrollRepository
	attendancePeriodRepo AttendancePeriodRepository
	uuidGen              pkguid.UUID
}

func NewAdminRunPayrollInteractor(
	payrollRepo PayrollRepository,
	attendancePeriodRepo AttendancePeriodRepository,
	uuidGen pkguid.UUID,
) AdminRunPayrollUsecase {
	return &adminRunPayrollInteractor{
		payrollRepo:          payrollRepo,
		attendancePeriodRepo: attendancePeriodRepo,
		uuidGen:              uuidGen,
	}
}

func (i *adminRunPayrollInteractor) Execute(ctx context.Context, in usecases.AdminRunPayrollInput) (usecases.AdminRunPayrollOutput, error) {
	// Get attendance period
	period, err := i.attendancePeriodRepo.GetByID(ctx, in.AttendancePeriodID)
	if err != nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewServerError("failed to get attendance period")
	}
	if period == nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewBusinessError("attendance period not found")
	}

	// Check if payroll already run
	existing, err := i.payrollRepo.GetPayrollRunByAttendancePeriod(ctx, sqlentity.StringToUUID(period.ID))
	if err != nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewServerError("failed to check existing payroll run")
	}
	if existing != nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewBusinessError("payroll already run for this period")
	}

	// Get all active employees
	employees, err := i.payrollRepo.GetActiveEmployees(ctx)
	if err != nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewServerError("failed to get active employees")
	}

	// Create payroll run
	payrollRun := &sqlentity.PayrollRuns{
		ID:                 i.uuidGen.Generate(),
		AttendancePeriodID: period.ID,
		RunAt:              time.Now(),
		CreatedAt:          time.Now(),
		CreatedBy:          sql.NullString{String: middleware.GetUserID(ctx), Valid: true},
	}

	// Save to database
	err = i.payrollRepo.CreatePayrollRun(ctx, payrollRun)
	if err != nil {
		return usecases.AdminRunPayrollOutput{}, pkgerror.NewServerError("failed to save payroll run")
	}

	return usecases.AdminRunPayrollOutput{
		PayrollRunID:       sqlentity.StringToUUID(payrollRun.ID),
		AttendancePeriodID: sqlentity.StringToUUID(payrollRun.AttendancePeriodID),
		RunAt:              payrollRun.RunAt,
		TotalEmployeeCount: len(employees),
	}, nil
}
