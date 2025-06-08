package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	EmployeeGeneratePayslipUsecase interface {
		Execute(ctx context.Context, input EmployeeGeneratePayslipInput) (EmployeeGeneratePayslipOutput, error)
	}

	EmployeeGeneratePayslipInput struct {
		EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
		Month      int       `json:"month" validate:"required,min=1,max=12"`
		Year       int       `json:"year" validate:"required"`
	}

	AttendanceSummary struct {
		Date  time.Time `json:"date"`
		Hours int       `json:"hours"`
	}

	OvertimeSummary struct {
		Date  time.Time `json:"date"`
		Hours int       `json:"hours"`
		Pay   float64   `json:"pay"`
	}

	ReimbursementSummary struct {
		Date        time.Time `json:"date"`
		Amount      float64   `json:"amount"`
		Description string    `json:"description"`
	}

	EmployeeGeneratePayslipOutput struct {
		ID                 uuid.UUID              `json:"id"`
		EmployeeID         uuid.UUID              `json:"employee_id"`
		EmployeeName       string                 `json:"employee_name"`
		Month              int                    `json:"month"`
		Year               int                    `json:"year"`
		BaseSalary         float64                `json:"base_salary"`
		AttendanceDays     int                    `json:"attendance_days"`
		AttendanceList     []AttendanceSummary    `json:"attendance_list"`
		OvertimeHours      int                    `json:"overtime_hours"`
		OvertimeList       []OvertimeSummary      `json:"overtime_list"`
		OvertimePay        float64                `json:"overtime_pay"`
		Reimbursements     []ReimbursementSummary `json:"reimbursements"`
		TotalReimbursement float64                `json:"total_reimbursement"`
		TakeHomePay        float64                `json:"take_home_pay"`
		GeneratedAt        time.Time              `json:"generated_at"`
	}
)
