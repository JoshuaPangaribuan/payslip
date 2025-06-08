package interactors

import (
	"context"
	"database/sql"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkguid"
	"github.com/shopspring/decimal"
)

// Usecase Interface
type EmployeeGeneratePayslipUsecase interface {
	Execute(ctx context.Context, in usecases.EmployeeGeneratePayslipInput) (usecases.EmployeeGeneratePayslipOutput, error)
}

// Repository Interface
type PayslipRepository interface {
	// CreatePayslip creates a new payslip
	CreatePayslip(ctx context.Context, payslip *sqlentity.Payslips) error
	// GetPayslipByEmployeeAndPeriod gets payslip by employee ID and period
	GetPayslipByEmployeeAndPeriod(ctx context.Context, employeeID string, month, year int) (*sqlentity.Payslips, error)
	// GetPayrollRunByMonthAndYear gets payroll run by month and year
	GetPayrollRunByMonthAndYear(ctx context.Context, month, year int) (*sqlentity.PayrollRuns, error)
	// GetEmployeeByID gets employee by ID
	GetEmployeeByID(ctx context.Context, id string) (*sqlentity.Employees, error)
	// GetAttendanceRecordsByEmployeeAndDateRange gets attendance records by employee ID and date range
	GetAttendanceRecordsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.AttendanceRecords, error)
	// GetOvertimeRequestsByEmployeeAndDateRange gets overtime requests by employee ID and date range
	GetOvertimeRequestsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.OvertimeRequests, error)
	// GetReimbursementRequestsByEmployeeAndDateRange gets reimbursement requests by employee ID and date range
	GetReimbursementRequestsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.ReimbursementRequests, error)
}

// Usecase Implementation
type employeeGeneratePayslipInteractor struct {
	payslipRepo PayslipRepository
	uuidGen     pkguid.UUID
}

func NewEmployeeGeneratePayslipInteractor(
	payslipRepo PayslipRepository,
	uuidGen pkguid.UUID,
) EmployeeGeneratePayslipUsecase {
	return &employeeGeneratePayslipInteractor{
		payslipRepo: payslipRepo,
		uuidGen:     uuidGen,
	}
}

func (i *employeeGeneratePayslipInteractor) Execute(ctx context.Context, in usecases.EmployeeGeneratePayslipInput) (usecases.EmployeeGeneratePayslipOutput, error) {
	// Get employee
	employee, err := i.payslipRepo.GetEmployeeByID(ctx, sqlentity.UUIDToString(in.EmployeeID))
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to get employee")
	}
	if employee == nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewBusinessError("employee not found")
	}

	// Get payroll run
	payrollRun, err := i.payslipRepo.GetPayrollRunByMonthAndYear(ctx, in.Month, in.Year)
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to get payroll run")
	}
	if payrollRun == nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewBusinessError("payroll run not found")
	}

	// Get attendance records
	attendanceRecords, err := i.payslipRepo.GetAttendanceRecordsByEmployeeAndDateRange(ctx, sqlentity.UUIDToString(in.EmployeeID), payrollRun.RunAt.AddDate(0, -1, 0), payrollRun.RunAt)
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to get attendance records")
	}

	// Get overtime requests
	overtimeRequests, err := i.payslipRepo.GetOvertimeRequestsByEmployeeAndDateRange(ctx, sqlentity.UUIDToString(in.EmployeeID), payrollRun.RunAt.AddDate(0, -1, 0), payrollRun.RunAt)
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to get overtime requests")
	}

	// Get reimbursement requests
	reimbursementRequests, err := i.payslipRepo.GetReimbursementRequestsByEmployeeAndDateRange(ctx, sqlentity.UUIDToString(in.EmployeeID), payrollRun.RunAt.AddDate(0, -1, 0), payrollRun.RunAt)
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to get reimbursement requests")
	}

	// Calculate total overtime pay
	totalOvertimePay := decimal.Zero
	for _, overtime := range overtimeRequests {
		hourlyRate := employee.SalaryMonthly.Div(decimal.NewFromInt(173)) // 173 working hours per month
		overtimePay := hourlyRate.Mul(decimal.NewFromInt(int64(overtime.Hours))).Mul(decimal.NewFromFloat(1.5))
		totalOvertimePay = totalOvertimePay.Add(overtimePay)
	}

	// Calculate total reimbursement
	totalReimbursement := decimal.Zero
	for _, reimbursement := range reimbursementRequests {
		totalReimbursement = totalReimbursement.Add(reimbursement.Amount)
	}

	// Calculate total take home pay
	totalTakeHomePay := employee.SalaryMonthly.Add(totalOvertimePay).Add(totalReimbursement)

	// Create payslip
	payslip := &sqlentity.Payslips{
		ID:                 i.uuidGen.Generate(),
		EmployeeID:         sqlentity.UUIDToString(in.EmployeeID),
		PayrollRunID:       payrollRun.ID,
		AttendancePay:      employee.SalaryMonthly,
		OvertimePay:        totalOvertimePay,
		ReimbursementTotal: totalReimbursement,
		TakeHomePay:        totalTakeHomePay,
		GeneratedAt:        time.Now(),
		CreatedAt:          time.Now(),
		CreatedBy:          sql.NullString{String: sqlentity.UUIDToString(in.EmployeeID), Valid: true},
	}

	// Save to database
	err = i.payslipRepo.CreatePayslip(ctx, payslip)
	if err != nil {
		return usecases.EmployeeGeneratePayslipOutput{}, pkgerror.NewServerError("failed to save payslip")
	}

	// Prepare attendance summary
	var attendanceList []usecases.AttendanceSummary
	for _, record := range attendanceRecords {
		attendanceList = append(attendanceList, usecases.AttendanceSummary{
			Date:  record.AttendanceDate,
			Hours: 8, // Assuming 8 hours per day
		})
	}

	// Prepare overtime summary
	var overtimeList []usecases.OvertimeSummary
	for _, overtime := range overtimeRequests {
		hourlyRate := employee.SalaryMonthly.Div(decimal.NewFromInt(173))
		overtimePay := hourlyRate.Mul(decimal.NewFromInt(int64(overtime.Hours))).Mul(decimal.NewFromFloat(1.5))
		overtimeList = append(overtimeList, usecases.OvertimeSummary{
			Date:  overtime.OvertimeDate,
			Hours: overtime.Hours,
			Pay:   overtimePay.InexactFloat64(),
		})
	}

	// Prepare reimbursement summary
	var reimbursements []usecases.ReimbursementSummary
	for _, reimbursement := range reimbursementRequests {
		reimbursements = append(reimbursements, usecases.ReimbursementSummary{
			Date:        reimbursement.CreatedAt,
			Amount:      reimbursement.Amount.InexactFloat64(),
			Description: reimbursement.Description.String,
		})
	}

	return usecases.EmployeeGeneratePayslipOutput{
		ID:                 sqlentity.StringToUUID(payslip.ID),
		EmployeeID:         sqlentity.StringToUUID(payslip.EmployeeID),
		EmployeeName:       employee.Username,
		Month:              in.Month,
		Year:               in.Year,
		BaseSalary:         payslip.AttendancePay.InexactFloat64(),
		AttendanceDays:     len(attendanceRecords),
		AttendanceList:     attendanceList,
		OvertimeHours:      len(overtimeRequests),
		OvertimeList:       overtimeList,
		OvertimePay:        payslip.OvertimePay.InexactFloat64(),
		Reimbursements:     reimbursements,
		TotalReimbursement: payslip.ReimbursementTotal.InexactFloat64(),
		TakeHomePay:        payslip.TakeHomePay.InexactFloat64(),
		GeneratedAt:        payslip.GeneratedAt,
	}, nil
}
