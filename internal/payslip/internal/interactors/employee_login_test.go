package interactors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	payslipmocks "github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/mocks"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockLogger struct {
	*zap.SugaredLogger
}

func TestEmployeeLoginInteractor_Execute(t *testing.T) {
	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	employee := &sqlentity.Employees{
		ID:             "emp-uuid",
		Username:       "johndoe",
		PasswordHashed: string(hashed),
		Role:           0,
	}

	tests := []struct {
		name          string
		input         usecases.EmployeeLoginInput
		mockEmployee  *sqlentity.Employees
		repoErr       error
		jwtSecret     string
		passwordInput string
		jwtSignErr    bool
		expectErr     bool
		assertToken   bool
	}{
		{
			name:          "valid login",
			input:         usecases.EmployeeLoginInput{Username: "johndoe", Password: password},
			mockEmployee:  employee,
			jwtSecret:     "secret",
			passwordInput: password,
			assertToken:   true,
		},
		{
			name:         "invalid username",
			input:        usecases.EmployeeLoginInput{Username: "notfound", Password: password},
			mockEmployee: nil,
			expectErr:    true,
		},
		{
			name:         "invalid password",
			input:        usecases.EmployeeLoginInput{Username: "johndoe", Password: "wrongpass"},
			mockEmployee: employee,
			expectErr:    true,
		},
		{
			name:      "repo error",
			input:     usecases.EmployeeLoginInput{Username: "johndoe", Password: password},
			repoErr:   errors.New("db error"),
			expectErr: true,
		},
		{
			name:         "jwt sign error",
			input:        usecases.EmployeeLoginInput{Username: "johndoe", Password: password},
			mockEmployee: employee,
			jwtSecret:    string([]byte{0xff, 0xfe, 0xfd}), // invalid secret for signing
			jwtSignErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(payslipmocks.MockEmployeeRepository)
			logger := zap.NewNop().Sugar()
			jwtSecret := tt.jwtSecret
			if jwtSecret == "" {
				jwtSecret = "secret"
			}

			if tt.repoErr != nil {
				repo.On("GetByUsername", mock.Anything, tt.input.Username).Return(nil, tt.repoErr)
			} else {
				repo.On("GetByUsername", mock.Anything, tt.input.Username).Return(tt.mockEmployee, nil)
			}

			interactor := NewEmployeeLoginInteractor(repo, jwtSecret, logger)
			output, err := interactor.Execute(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, output.Token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, output.Token)
				// Optionally, parse the token to check claims
				token, parseErr := jwt.Parse(output.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				assert.NoError(t, parseErr)
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					assert.Equal(t, employee.ID, claims["sub"])
					assert.Equal(t, float64(employee.Role), claims["role"])
					assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), time.Unix(int64(claims["exp"].(float64)), 0), time.Minute*2)
				}
			}
			repo.AssertExpectations(t)
		})
	}
}
