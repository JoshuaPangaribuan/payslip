package interactors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/middleware"
	payslipmocks "github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/mocks"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	pkgmocks "github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgmocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestEmployeeSubmitOvertimeInteractor_Execute(t *testing.T) {
	now := time.Date(2025, 3, 17, 10, 0, 0, 0, time.UTC) // Monday
	periodID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	employeeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	overtimeID := "33333333-3333-3333-3333-333333333333"
	auditID := "44444444-4444-4444-4444-444444444444"

	period := &sqlentity.AttendancePeriods{
		ID:        periodID.String(),
		StartDate: now.Add(-24 * time.Hour),
		EndDate:   now.Add(24 * time.Hour),
	}

	tests := []struct {
		name          string
		input         usecases.EmployeeSubmitOvertimeInput
		ctxUserID     string
		mockPeriod    *sqlentity.AttendancePeriods
		periodErr     error
		mockExisting  *sqlentity.OvertimeRequests
		existingErr   error
		createErr     error
		auditErr      error
		expectErr     bool
		expectOutput  bool
		expectedError string
	}{
		{
			name: "valid overtime submission",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:    employeeID.String(),
			mockPeriod:   period,
			expectOutput: true,
		},
		{
			name: "invalid hours - too low",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              0,
			},
			ctxUserID:     employeeID.String(),
			expectErr:     true,
			expectedError: "overtime hours must be between 1 and 3",
		},
		{
			name: "invalid hours - too high",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              4,
			},
			ctxUserID:     employeeID.String(),
			expectErr:     true,
			expectedError: "overtime hours must be between 1 and 3",
		},
		{
			name: "date outside period",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now.Add(-48 * time.Hour),
				Hours:              2,
			},
			ctxUserID:     employeeID.String(),
			mockPeriod:    period,
			expectErr:     true,
			expectedError: "date must be within attendance period range",
		},
		{
			name: "overtime already exists",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:     employeeID.String(),
			mockPeriod:    period,
			mockExisting:  &sqlentity.OvertimeRequests{ID: overtimeID},
			expectErr:     true,
			expectedError: "overtime already exists for this date",
		},
		{
			name: "period not found",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:     employeeID.String(),
			mockPeriod:    nil,
			expectErr:     true,
			expectedError: "attendance period not found",
		},
		{
			name: "repo error",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:     employeeID.String(),
			periodErr:     errors.New("db error"),
			expectErr:     true,
			expectedError: "failed to get attendance period",
		},
		{
			name: "audit log error",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:    employeeID.String(),
			mockPeriod:   period,
			auditErr:     errors.New("audit error"),
			expectOutput: true,
		},
		{
			name: "unauthorized",
			input: usecases.EmployeeSubmitOvertimeInput{
				AttendancePeriodID: periodID,
				Date:               now,
				Hours:              2,
			},
			ctxUserID:     "",
			expectErr:     true,
			expectedError: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overtimeRepo := new(payslipmocks.MockOvertimeRepository)
			auditRepo := new(payslipmocks.MockAuditLogRepository)
			periodRepo := new(payslipmocks.MockAttendancePeriodRepository)
			uuidGen := new(pkgmocks.MockUUID)
			logger := zap.NewNop().Sugar()

			shouldCreate := tt.name == "valid overtime submission" || tt.name == "audit log error"
			if shouldCreate {
				uuidGen.On("Generate").Return(overtimeID).Once()
				uuidGen.On("Generate").Return(auditID).Once()
				overtimeRepo.On("CreateOvertime", mock.Anything, mock.AnythingOfType("*sqlentity.OvertimeRequests")).Return(tt.createErr)
				auditRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*sqlentity.AuditLogs")).Return(tt.auditErr)
				overtimeRepo.On("GetOvertimeByEmployeeAndDate", mock.Anything, employeeID, tt.input.Date).Return(nil, nil)
			}

			if tt.name != "unauthorized" && tt.name != "invalid hours - too low" && tt.name != "invalid hours - too high" {
				periodRepo.On("GetByID", mock.Anything, tt.input.AttendancePeriodID).Return(tt.mockPeriod, tt.periodErr)
			}

			if tt.mockExisting != nil || tt.existingErr != nil {
				overtimeRepo.On("GetOvertimeByEmployeeAndDate", mock.Anything, employeeID, tt.input.Date).Return(tt.mockExisting, tt.existingErr)
			}

			ctx := context.Background()
			if tt.ctxUserID != "" {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.ctxUserID)
			}

			interactor := NewEmployeeSubmitOvertimeInteractor(overtimeRepo, auditRepo, uuidGen, logger, periodRepo)
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
					assert.Equal(t, uuid.MustParse(overtimeID), output.ID)
					assert.Equal(t, uuid.MustParse(employeeID.String()), output.EmployeeID)
					assert.Equal(t, tt.input.Date, output.Date)
					assert.Equal(t, tt.input.Hours, output.Hours)
					assert.WithinDuration(t, time.Now(), output.CreatedAt, time.Second*2)
				}
			}

			overtimeRepo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
			periodRepo.AssertExpectations(t)
			uuidGen.AssertExpectations(t)
		})
	}
}
