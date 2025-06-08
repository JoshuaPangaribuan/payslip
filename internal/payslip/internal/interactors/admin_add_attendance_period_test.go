package interactors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	payslipmocks "github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/mocks"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	pkgmocks "github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestAdminAddAttendancePeriodInteractor_Execute(t *testing.T) {
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)
	periodID := "11111111-1111-1111-1111-111111111111"
	auditID := "22222222-2222-2222-2222-222222222222"
	userID := "admin-uuid"

	tests := []struct {
		name         string
		input        usecases.AdminAddAttendancePeriodInput
		mockExisting *sqlentity.AttendancePeriods
		existingErr  error
		createErr    error
		auditErr     error
		expectErr    bool
		expectOutput bool
	}{
		{
			name:         "valid period",
			input:        usecases.AdminAddAttendancePeriodInput{StartDate: start, EndDate: end},
			expectOutput: true,
		},
		{
			name:      "end before start",
			input:     usecases.AdminAddAttendancePeriodInput{StartDate: end, EndDate: start},
			expectErr: true,
		},
		{
			name:         "period already exists",
			input:        usecases.AdminAddAttendancePeriodInput{StartDate: start, EndDate: end},
			mockExisting: &sqlentity.AttendancePeriods{ID: periodID, StartDate: start, EndDate: end},
			expectErr:    true,
		},
		{
			name:        "repo error",
			input:       usecases.AdminAddAttendancePeriodInput{StartDate: start, EndDate: end},
			existingErr: errors.New("db error"),
			createErr:   errors.New("db error"),
			expectErr:   true,
		},
		{
			name:         "audit log error",
			input:        usecases.AdminAddAttendancePeriodInput{StartDate: start, EndDate: end},
			auditErr:     errors.New("audit error"),
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(payslipmocks.MockAttendancePeriodRepository)
			auditRepo := new(payslipmocks.MockAuditLogRepository)
			uuidGen := new(pkgmocks.MockUUID)
			logger := zap.NewNop().Sugar()

			shouldCreate := tt.name == "valid period" || tt.name == "audit log error"
			if shouldCreate {
				uuidGen.On("Generate").Return(periodID).Once()
				uuidGen.On("Generate").Return(auditID).Once()
				repo.On("CreateAttendancePeriod", mock.Anything, mock.AnythingOfType("*sqlentity.AttendancePeriods")).Return(tt.createErr)
				auditRepo.On("CreateAuditLog", mock.Anything, mock.AnythingOfType("*sqlentity.AuditLogs")).Return(tt.auditErr)
			}

			if tt.name == "repo error" {
				uuidGen.On("Generate").Return(periodID).Once()
				repo.On("CreateAttendancePeriod", mock.Anything, mock.AnythingOfType("*sqlentity.AttendancePeriods")).Return(tt.createErr)
			}

			// Only set up mocks if the code is expected to reach those calls
			if tt.name != "end before start" {
				if tt.mockExisting != nil || tt.existingErr != nil {
					repo.On("GetAttendancePeriodByDateRange", mock.Anything, tt.input.StartDate, tt.input.EndDate).Return(tt.mockExisting, tt.existingErr)
				} else {
					repo.On("GetAttendancePeriodByDateRange", mock.Anything, tt.input.StartDate, tt.input.EndDate).Return(nil, nil)
				}
			}

			// Set user ID in context
			ctx := context.WithValue(context.Background(), "user_id", userID)

			interactor := NewAdminAddAttendancePeriodInteractor(repo, auditRepo, uuidGen, logger)
			output, err := interactor.Execute(ctx, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, output.ID)
			} else {
				assert.NoError(t, err)
				if tt.expectOutput {
					assert.Equal(t, periodID, output.ID.String())
					assert.Equal(t, tt.input.StartDate, output.StartDate)
					assert.Equal(t, tt.input.EndDate, output.EndDate)
					assert.WithinDuration(t, time.Now(), output.CreatedAt, time.Second*2)
				}
			}
			repo.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
			uuidGen.AssertExpectations(t)
		})
	}
}
