package interactors

import (
	"context"
	"errors"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Usecase Interface
type AdminPayrollSummaryUsecase interface {
	Execute(ctx context.Context, in usecases.AdminPayrollSummaryInput) (usecases.AdminPayrollSummaryOutput, error)
}

// Repository Interface
type PayrollSummaryRepository interface {
	// GetPayrollRun gets payroll run by ID
	GetPayrollRun(ctx context.Context, payrollRunID uuid.UUID) (*sqlentity.PayrollRuns, error)
	// GetAttendancePeriod gets attendance period by ID
	GetAttendancePeriod(ctx context.Context, attendancePeriodID uuid.UUID) (*sqlentity.AttendancePeriods, error)
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
type adminPayrollSummaryInteractor struct {
	payrollSummaryRepo PayrollSummaryRepository
}

func NewAdminPayrollSummaryInteractor(
	payrollSummaryRepo PayrollSummaryRepository,
) AdminPayrollSummaryUsecase {
	return &adminPayrollSummaryInteractor{
		payrollSummaryRepo: payrollSummaryRepo,
	}
}

func (i *adminPayrollSummaryInteractor) Execute(ctx context.Context, in usecases.AdminPayrollSummaryInput) (usecases.AdminPayrollSummaryOutput, error) {
	// Get payroll run
	payrollRun, err := i.payrollSummaryRepo.GetPayrollRun(ctx, in.PayrollRunID)
	if err != nil {
		return usecases.AdminPayrollSummaryOutput{}, errors.New("payroll run not found")
	}

	// Get attendance period
	attendancePeriod, err := i.payrollSummaryRepo.GetAttendancePeriod(ctx, sqlentity.StringToUUID(payrollRun.AttendancePeriodID))
	if err != nil {
		return usecases.AdminPayrollSummaryOutput{}, errors.New("attendance period not found")
	}

	// Get all active employees
	employees, err := i.payrollSummaryRepo.GetActiveEmployees(ctx)
	if err != nil {
		return usecases.AdminPayrollSummaryOutput{}, err
	}

	// Calculate summary for each employee
	var employeeSummaries []usecases.EmployeePayrollSummary
	var totalTakeHomePay decimal.Decimal

	for _, employee := range employees {
		// Get attendance records
		attendanceRecords, err := i.payrollSummaryRepo.GetAttendanceByPeriod(ctx, sqlentity.StringToUUID(employee.ID), attendancePeriod.StartDate, attendancePeriod.EndDate)
		if err != nil {
			return usecases.AdminPayrollSummaryOutput{}, err
		}

		// Get overtime requests
		overtimeRequests, err := i.payrollSummaryRepo.GetOvertimeByPeriod(ctx, sqlentity.StringToUUID(employee.ID), attendancePeriod.StartDate, attendancePeriod.EndDate)
		if err != nil {
			return usecases.AdminPayrollSummaryOutput{}, err
		}

		// Get reimbursement requests
		reimbursementRequests, err := i.payrollSummaryRepo.GetReimbursementsByPeriod(ctx, sqlentity.StringToUUID(employee.ID), attendancePeriod.StartDate, attendancePeriod.EndDate)
		if err != nil {
			return usecases.AdminPayrollSummaryOutput{}, err
		}

		// Calculate base salary (prorated based on attendance)
		baseSalary := employee.SalaryMonthly.Mul(decimal.NewFromInt(int64(len(attendanceRecords)))).Div(decimal.NewFromInt(22)) // 22 working days per month

		// Calculate overtime pay
		var overtimePay decimal.Decimal
		for _, overtime := range overtimeRequests {
			hourlyRate := employee.SalaryMonthly.Div(decimal.NewFromInt(176))                                                   // 176 working hours per month (22 days * 8 hours)
			overtimePay = overtimePay.Add(hourlyRate.Mul(decimal.NewFromInt(int64(overtime.Hours))).Mul(decimal.NewFromInt(2))) // 2x rate for overtime
		}

		// Calculate reimbursements
		var reimbursements decimal.Decimal
		for _, reimbursement := range reimbursementRequests {
			reimbursements = reimbursements.Add(reimbursement.Amount)
		}

		// Calculate take home pay
		takeHomePay := baseSalary.Add(overtimePay).Add(reimbursements)
		totalTakeHomePay = totalTakeHomePay.Add(takeHomePay)

		// Add to summaries
		employeeSummaries = append(employeeSummaries, usecases.EmployeePayrollSummary{
			EmployeeID:     sqlentity.StringToUUID(employee.ID),
			EmployeeName:   employee.Username,
			BaseSalary:     baseSalary,
			OvertimePay:    overtimePay,
			Reimbursements: reimbursements,
			TakeHomePay:    takeHomePay,
		})
	}

	return usecases.AdminPayrollSummaryOutput{
		PayrollRunID:       sqlentity.StringToUUID(payrollRun.ID),
		AttendancePeriodID: sqlentity.StringToUUID(attendancePeriod.ID),
		ProcessedAt:        payrollRun.RunAt,
		EmployeeSummaries:  employeeSummaries,
		TotalTakeHomePay:   totalTakeHomePay,
	}, nil
}
