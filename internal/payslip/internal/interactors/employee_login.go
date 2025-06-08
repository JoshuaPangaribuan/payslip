package interactors

import (
	"context"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Usecase Interface
type EmployeeLoginUsecase interface {
	Execute(ctx context.Context, in usecases.EmployeeLoginInput) (usecases.EmployeeLoginOutput, error)
}

// Repository Interface
type EmployeeRepository interface {
	// GetByUsername retrieves an employee by their username
	GetByUsername(ctx context.Context, username string) (*sqlentity.Employees, error)
}

// Usecase Implementation
type employeeLoginInteractor struct {
	employeeRepo EmployeeRepository
	jwtSecret    string
	logger       *zap.SugaredLogger
}

func NewEmployeeLoginInteractor(
	employeeRepo EmployeeRepository,
	jwtSecret string,
	logger *zap.SugaredLogger,
) EmployeeLoginUsecase {
	return &employeeLoginInteractor{
		employeeRepo: employeeRepo,
		jwtSecret:    jwtSecret,
		logger:       logger,
	}
}

func (i *employeeLoginInteractor) Execute(ctx context.Context, in usecases.EmployeeLoginInput) (usecases.EmployeeLoginOutput, error) {
	i.logger.Infow("starting employee login attempt",
		"username", in.Username,
	)

	// Get employee by username
	employee, err := i.employeeRepo.GetByUsername(ctx, in.Username)
	if err != nil {
		i.logger.Errorw("failed to get employee by username",
			"error", err,
			"username", in.Username,
		)
		return usecases.EmployeeLoginOutput{}, pkgerror.NewBusinessError("invalid username or password")
	}

	if employee == nil {
		i.logger.Warnw("employee not found",
			"username", in.Username,
		)
		return usecases.EmployeeLoginOutput{}, pkgerror.NewBusinessError("invalid username or password")
	}

	i.logger.Infow("employee found",
		"employee_id", employee.ID,
		"username", employee.Username,
		"role", employee.Role,
	)

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(employee.PasswordHashed), []byte(in.Password))
	if err != nil {
		i.logger.Warnw("invalid password",
			"employee_id", employee.ID,
			"username", employee.Username,
		)
		return usecases.EmployeeLoginOutput{}, pkgerror.NewBusinessError("invalid username or password")
	}

	i.logger.Infow("password verified successfully",
		"employee_id", employee.ID,
		"username", employee.Username,
	)

	// Generate JWT token
	expiryTime := time.Now().Add(7 * (24 * time.Hour))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  employee.ID,
		"role": employee.Role,
		"exp":  expiryTime.Unix(),
	})

	i.logger.Infow("generating JWT token",
		"employee_id", employee.ID,
		"role", employee.Role,
		"expiry_time", expiryTime,
	)

	// Sign token
	tokenString, err := token.SignedString([]byte(i.jwtSecret))
	if err != nil {
		i.logger.Errorw("failed to sign JWT token",
			"error", err,
			"employee_id", employee.ID,
		)
		return usecases.EmployeeLoginOutput{}, pkgerror.NewServerError("failed to generate token")
	}

	i.logger.Infow("login successful",
		"employee_id", employee.ID,
		"username", employee.Username,
		"role", employee.Role,
		"token_expiry", expiryTime,
	)

	return usecases.EmployeeLoginOutput{
		Token: tokenString,
	}, nil
}
