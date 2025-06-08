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

func TestEmployeeSubmitAttendanceInteractor_Execute(t *testing.T) {
	now := time.Date(2025, 3, 17, 10, 0, 0, 0, time.UTC)     // Monday
	weekend := time.Date(2025, 3, 16, 10, 0, 0, 0, time.UTC) // Sunday
	periodID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	employeeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	attendanceID := "33333333-3333-3333-3333-333333333333"
	auditID := "44444444-4444-4444-4444-444444444444"

	period := &sqlentity.AttendancePeriods{
		ID:        periodID.String(),
		StartDate: now.Add(-24 * time.Hour),
		EndDate:   now.Add(24 * time.Hour),
	}

	tests := []struct {
		name         string
		input        usecases.EmployeeSubmitAttendanceInput
		ctxUserID    string
		mockPeriod   *sqlentity.AttendancePeriods
		periodErr    error
		mockExisting *sqlentity.AttendanceRecords
		existingErr  error
		createErr    error
		auditErr     error
		expectErr    bool
		expectOutput bool
	}{
		{
			name:         "valid attendance",
			input:        usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID:    employeeID.String(),
			mockPeriod:   period,
			expectOutput: true,
		},
		{
			name:      "weekend",
			input:     usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: weekend},
			ctxUserID: employeeID.String(),
			expectErr: true,
		},
		{
			name:       "date outside period",
			input:      usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now.Add(-48 * time.Hour)},
			ctxUserID:  employeeID.String(),
			mockPeriod: period,
			periodErr:  nil,
			expectErr:  true,
		},
		{
			name:         "attendance already exists",
			input:        usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID:    employeeID.String(),
			mockPeriod:   period,
			mockExisting: &sqlentity.AttendanceRecords{ID: attendanceID},
			expectErr:    true,
		},
		{
			name:       "period not found",
			input:      usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID:  employeeID.String(),
			mockPeriod: nil,
			periodErr:  nil,
			expectErr:  true,
		},
		{
			name:      "repo error",
			input:     usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID: employeeID.String(),
			periodErr: errors.New("db error"),
			expectErr: true,
		},
		{
			name:         "audit log error",
			input:        usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID:    employeeID.String(),
			mockPeriod:   period,
			auditErr:     errors.New("audit error"),
			expectOutput: true,
		},
		{
			name:      "unauthorized",
			input:     usecases.EmployeeSubmitAttendanceInput{AttendancePeriodID: periodID, Date: now},
			ctxUserID: "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(payslipmocks.MockAttendanceRepository)
			auditRepo := new(payslipmocks.MockAuditLogRepository)
			uuidGen := new(pkgmocks.MockUUID)
			logger := zap.NewNop().Sugar()

			shouldCreate := tt.name == "valid attendance" || tt.name == "audit log error"
			if shouldCreate {
				uuidGen.On("Generate").Return(attendanceID).Once()
				uuidGen.On("Generate").Return(auditID).Once()
				repo.On("CreateAttendance", mock.Anything, mock.AnythingOfType("*sqlentity.AttendanceRecords")).Return(tt.createErr)
				auditRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*sqlentity.AuditLogs")).Return(tt.auditErr)
				repo.On("GetAttendanceByEmployeeAndDate", mock.Anything, mock.Anything, tt.input.Date).Return(nil, nil)
			}

			if tt.name != "weekend" && tt.name != "unauthorized" {
				repo.On("GetAttendancePeriodByID", mock.Anything, tt.input.AttendancePeriodID).Return(tt.mockPeriod, tt.periodErr)
			}
			if tt.mockExisting != nil || tt.existingErr != nil {
				repo.On("GetAttendanceByEmployeeAndDate", mock.Anything, mock.Anything, tt.input.Date).Return(tt.mockExisting, tt.existingErr)
			}

			ctx := context.Background()
			if tt.ctxUserID != "" {
				ctx = context.WithValue(ctx, middleware.UserIDKey, tt.ctxUserID)
			}

			interactor := NewEmployeeSubmitAttendanceInteractor(repo, auditRepo, uuidGen, logger)
			output, err := interactor.Execute(ctx, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, output.ID)
			} else {
				assert.NoError(t, err)
				if tt.expectOutput {
					assert.Equal(t, uuid.MustParse(attendanceID), output.ID)
					assert.Equal(t, uuid.MustParse(employeeID.String()), output.EmployeeID)
					assert.Equal(t, tt.input.Date, output.Date)
					assert.WithinDuration(t, time.Now(), output.CreatedAt, time.Second*2)
				}
			}
			// repo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
			uuidGen.AssertExpectations(t)
		})
	}
}
