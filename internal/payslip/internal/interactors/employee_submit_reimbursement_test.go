package interactors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/middleware"
	payslipmocks "github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/mocks"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	pkgmocks "github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgmocks"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestEmployeeSubmitReimbursementInteractor_Execute(t *testing.T) {
	periodID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	employeeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	reimbursementID := "33333333-3333-3333-3333-333333333333"
	auditID := "44444444-4444-4444-4444-444444444444"

	tests := []struct {
		name          string
		input         usecases.EmployeeSubmitReimbursementInput
		ctxUserID     string
		createErr     error
		auditErr      error
		expectErr     bool
		expectOutput  bool
		expectedError string
	}{
		{
			name: "valid reimbursement submission",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.NewFromInt(100000),
				Description:        "Medical claim",
			},
			ctxUserID:    employeeID.String(),
			expectOutput: true,
		},
		{
			name: "amount zero",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.Zero,
				Description:        "Transport",
			},
			ctxUserID:     employeeID.String(),
			expectErr:     true,
			expectedError: "amount must be greater than 0",
		},
		{
			name: "missing attendance period",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: uuid.Nil,
				Amount:             decimal.NewFromInt(50000),
				Description:        "Meal",
			},
			ctxUserID:     employeeID.String(),
			expectErr:     true,
			expectedError: "attendance period ID is required",
		},
		{
			name: "unauthorized",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.NewFromInt(10000),
				Description:        "Other",
			},
			ctxUserID:     "",
			expectErr:     true,
			expectedError: "unauthorized",
		},
		{
			name: "invalid employee ID",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.NewFromInt(10000),
				Description:        "Other",
			},
			ctxUserID:     "not-a-uuid",
			expectErr:     true,
			expectedError: "invalid employee ID",
		},
		{
			name: "repo error",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.NewFromInt(10000),
				Description:        "Other",
			},
			ctxUserID:     employeeID.String(),
			createErr:     errors.New("db error"),
			expectErr:     true,
			expectedError: "failed to save reimbursement",
		},
		{
			name: "audit log error",
			input: usecases.EmployeeSubmitReimbursementInput{
				EmployeeID:         employeeID,
				AttendancePeriodID: periodID,
				Amount:             decimal.NewFromInt(10000),
				Description:        "Other",
			},
			ctxUserID:    employeeID.String(),
			auditErr:     errors.New("audit error"),
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reimbursementRepo := new(payslipmocks.MockReimbursementRepository)
			auditRepo := new(payslipmocks.MockAuditLogRepository)
			uuidGen := new(pkgmocks.MockUUID)
			logger := zap.NewNop().Sugar()

			shouldCreate := tt.name == "valid reimbursement submission" || tt.name == "audit log error"
			if shouldCreate {
				uuidGen.On("Generate").Return(reimbursementID).Once()
				uuidGen.On("Generate").Return(auditID).Once()
				reimbursementRepo.On("CreateReimbursement", mock.Anything, mock.AnythingOfType("*sqlentity.ReimbursementRequests")).Return(tt.createErr)
				auditRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*sqlentity.AuditLogs")).Return(tt.auditErr)
			}

			if tt.name == "repo error" {
				uuidGen.On("Generate").Return(reimbursementID).Once()
				reimbursementRepo.On("CreateReimbursement", mock.Anything, mock.AnythingOfType("*sqlentity.ReimbursementRequests")).Return(tt.createErr)
			}

			ctx := context.Background()
			if tt.ctxUserID != "" {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.ctxUserID)
			}

			interactor := NewEmployeeSubmitReimbursementInteractor(reimbursementRepo, auditRepo, logger, uuidGen)
			output, err := interactor.Execute(ctx, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, output.ID)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectOutput {
					assert.Equal(t, uuid.MustParse(reimbursementID), output.ID)
					assert.Equal(t, employeeID, output.EmployeeID)
					assert.Equal(t, periodID, output.AttendancePeriodID)
					assert.Equal(t, tt.input.Amount, output.Amount)
					assert.Equal(t, tt.input.Description, output.Description)
					assert.WithinDuration(t, time.Now(), output.CreatedAt, time.Second*2)
				}
			}

			reimbursementRepo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
			uuidGen.AssertExpectations(t)
		})
	}
}
