package usecases

import "context"

type (
	EmployeeLoginUsecase interface {
		Execute(ctx context.Context, in EmployeeLoginInput) (EmployeeLoginOutput, error)
	}

	EmployeeLoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	EmployeeLoginOutput struct {
		Token string `json:"token"`
	}
)
