package gateways

import (
	"context"
	"net/http"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/middleware"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/usecases"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkghttp/v1"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/shopspring/decimal"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title           Payslip API
// @version         1.0
// @description     A payslip generation system API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8081
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

const (
	// Common endpoints
	LoginURL = "/api/v1/login"

	// Admin endpoints
	AdminAttendancePeriodURL = "/api/v1/admin/attendance-period"
	AdminPayrollRunURL       = "/api/v1/admin/payroll-run"
	AdminPayrollSummaryURL   = "/api/v1/admin/payroll-summary"

	// Employee endpoints
	EmployeeAttendanceURL      = "/api/v1/employee/attendance"
	EmployeeOvertimeURL        = "/api/v1/employee/overtime"
	EmployeeReimbursementURL   = "/api/v1/employee/reimbursement"
	EmployeeGeneratePayslipURL = "/api/v1/employee/generate-payslip"

	// Swagger endpoint
	SwaggerURL = "/swagger/*any"
)

func NewPayslipHttpGateway(
	router *httprouter.Router,
	logger *zap.SugaredLogger,
	validator *validator.Validate,
	adminEndpoint *AdminHTTPEndpoint,
	employeeEndpoint *EmployeeHTTPEndpoint,
	commonEndpoint *CommonHTTPEndpoint,
) {
	// Add Swagger documentation endpoint
	router.Handler(http.MethodGet, SwaggerURL, httpSwagger.WrapHandler)

	// Common endpoints (no auth required)
	commonServer := pkghttp.NewServer(
		pkghttp.WithResponseEncoder(pkghttp.DefaultResponseEncoder),
		pkghttp.WithErrorResponseEncoder(pkghttp.DefaultErrorEncoder),
	)

	router.Handler(
		http.MethodPost,
		LoginURL,
		middleware.WithRequestID(
			middleware.WithIPAddress(commonServer.Serve(commonEndpoint.Login)),
		),
	)

	// Admin endpoints (require JWT + Admin role)
	adminServer := pkghttp.NewServer(
		pkghttp.WithResponseEncoder(pkghttp.DefaultResponseEncoder),
		pkghttp.WithErrorResponseEncoder(pkghttp.DefaultErrorEncoder),
	)

	// Apply JWT middleware to all admin routes
	adminHandler := middleware.WithRequestID(
		middleware.WithIPAddress(
			middleware.JWTMiddleware(
				middleware.AdminOnly(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Route to appropriate handler based on path
						switch r.URL.Path {
						case AdminAttendancePeriodURL:
							adminServer.Serve(adminEndpoint.AddAttendancePeriod).ServeHTTP(w, r)
						case AdminPayrollRunURL:
							adminServer.Serve(adminEndpoint.RunPayroll).ServeHTTP(w, r)
						case AdminPayrollSummaryURL:
							adminServer.Serve(adminEndpoint.PayrollSummary).ServeHTTP(w, r)
						}
					}),
				),
			),
		),
	)

	router.Handler(
		http.MethodPost,
		AdminAttendancePeriodURL,
		adminHandler,
	)

	router.Handler(
		http.MethodPost,
		AdminPayrollRunURL,
		adminHandler,
	)

	router.Handler(
		http.MethodGet,
		AdminPayrollSummaryURL,
		adminHandler,
	)

	// Employee endpoints (require JWT + Employee role)
	employeeServer := pkghttp.NewServer(
		pkghttp.WithResponseEncoder(pkghttp.DefaultResponseEncoder),
		pkghttp.WithErrorResponseEncoder(pkghttp.DefaultErrorEncoder),
	)

	// Apply JWT middleware to all employee routes
	employeeHandler := middleware.WithRequestID(
		middleware.WithIPAddress(
			middleware.JWTMiddleware(
				middleware.EmployeeOnly(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Route to appropriate handler based on path
						switch r.URL.Path {
						case EmployeeAttendanceURL:
							employeeServer.Serve(employeeEndpoint.SubmitAttendance).ServeHTTP(w, r)
						case EmployeeOvertimeURL:
							employeeServer.Serve(employeeEndpoint.Overtime).ServeHTTP(w, r)
						case EmployeeReimbursementURL:
							employeeServer.Serve(employeeEndpoint.Reimbursement).ServeHTTP(w, r)
						case EmployeeGeneratePayslipURL:
							employeeServer.Serve(employeeEndpoint.GeneratePayslip).ServeHTTP(w, r)
						}
					}),
				),
			),
		),
	)

	router.Handler(
		http.MethodPost,
		EmployeeAttendanceURL,
		employeeHandler,
	)

	router.Handler(
		http.MethodPost,
		EmployeeOvertimeURL,
		employeeHandler,
	)

	router.Handler(
		http.MethodPost,
		EmployeeReimbursementURL,
		employeeHandler,
	)

	router.Handler(
		http.MethodGet,
		EmployeeGeneratePayslipURL,
		employeeHandler,
	)
}

// Common HTTP Endpoint
type CommonHTTPEndpoint struct {
	employeeLoginUsecase usecases.EmployeeLoginUsecase

	logger    *zap.SugaredLogger
	validator *validator.Validate
}

func NewCommonHTTPEndpoint(
	employeeLoginUsecase usecases.EmployeeLoginUsecase,
	logger *zap.SugaredLogger,
	validator *validator.Validate,
) *CommonHTTPEndpoint {
	return &CommonHTTPEndpoint{
		employeeLoginUsecase,
		logger,
		validator,
	}
}

// @Summary      Employee login
// @Description  Authenticate employee and get JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body usecases.EmployeeLoginInput true "Login credentials"
// @Success      200  {object}  usecases.EmployeeLoginOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Router       /login [post]
func (che *CommonHTTPEndpoint) Login(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.EmployeeLoginInput
	if err := request.Decode(&input); err != nil {
		che.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := che.validator.Struct(input); err != nil {
		che.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := che.employeeLoginUsecase.Execute(ctx, input)
	if err != nil {
		che.logger.Errorw("failed to execute employee login usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// Admin HTTP Endpoint
type AdminHTTPEndpoint struct {
	adminAddAttendancePeriodUsecase usecases.AdminAddAttendancePeriodUsecase
	adminRunPayrollUsecase          usecases.AdminRunPayrollUsecase
	adminPayrollSummaryUsecase      usecases.AdminPayrollSummaryUsecase

	logger    *zap.SugaredLogger
	validator *validator.Validate
}

func NewAdminHTTPEndpoint(
	adminAddAttendancePeriodUsecase usecases.AdminAddAttendancePeriodUsecase,
	adminRunPayrollUsecase usecases.AdminRunPayrollUsecase,
	adminPayrollSummaryUsecase usecases.AdminPayrollSummaryUsecase,
	logger *zap.SugaredLogger,
	validator *validator.Validate,
) *AdminHTTPEndpoint {
	return &AdminHTTPEndpoint{
		adminAddAttendancePeriodUsecase,
		adminRunPayrollUsecase,
		adminPayrollSummaryUsecase,
		logger,
		validator,
	}
}

// @Summary      Add attendance period
// @Description  Create a new attendance period for employees to submit their attendance
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        request body usecases.AdminAddAttendancePeriodInput true "Attendance period details"
// @Success      200  {object}  usecases.AdminAddAttendancePeriodOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /admin/attendance-period [post]
func (ahttp *AdminHTTPEndpoint) AddAttendancePeriod(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.AdminAddAttendancePeriodInput
	if err := request.Decode(&input); err != nil {
		ahttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ahttp.validator.Struct(input); err != nil {
		ahttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ahttp.adminAddAttendancePeriodUsecase.Execute(ctx, input)
	if err != nil {
		ahttp.logger.Errorw("failed to execute admin add attendance period usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// @Summary      Run payroll
// @Description  Process payroll for a specific period
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        request body usecases.AdminRunPayrollInput true "Payroll run details"
// @Success      200  {object}  usecases.AdminRunPayrollOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /admin/payroll-run [post]
func (ahttp *AdminHTTPEndpoint) RunPayroll(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.AdminRunPayrollInput
	if err := request.Decode(&input); err != nil {
		ahttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ahttp.validator.Struct(input); err != nil {
		ahttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ahttp.adminRunPayrollUsecase.Execute(ctx, input)
	if err != nil {
		ahttp.logger.Errorw("failed to execute admin run payroll usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// @Summary      Get payroll summary
// @Description  Get summary of all payslips for a specific period
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param        request body usecases.AdminPayrollSummaryInput true "Payroll summary details"
// @Success      200  {object}  usecases.AdminPayrollSummaryOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /admin/payroll-summary [get]
func (ahttp *AdminHTTPEndpoint) PayrollSummary(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.AdminPayrollSummaryInput
	if err := request.Decode(&input); err != nil {
		ahttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ahttp.validator.Struct(input); err != nil {
		ahttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ahttp.adminPayrollSummaryUsecase.Execute(ctx, input)
	if err != nil {
		ahttp.logger.Errorw("failed to execute admin payroll summary usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// Employee HTTP Endpoint
type EmployeeHTTPEndpoint struct {
	employeeSubmitAttendanceUsecase    usecases.EmployeeSubmitAttendanceUsecase
	employeeSubmitOvertimeUsecase      usecases.EmployeeSubmitOvertimeUsecase
	employeeSubmitReimbursementUsecase usecases.EmployeeSubmitReimbursementUsecase
	employeeGeneratePayslipUsecase     usecases.EmployeeGeneratePayslipUsecase

	logger    *zap.SugaredLogger
	validator *validator.Validate
}

func NewEmployeeHTTPEndpoint(
	employeeSubmitAttendanceUsecase usecases.EmployeeSubmitAttendanceUsecase,
	employeeSubmitOvertimeUsecase usecases.EmployeeSubmitOvertimeUsecase,
	employeeSubmitReimbursementUsecase usecases.EmployeeSubmitReimbursementUsecase,
	employeeGeneratePayslipUsecase usecases.EmployeeGeneratePayslipUsecase,
	logger *zap.SugaredLogger,
	validator *validator.Validate,
) *EmployeeHTTPEndpoint {
	return &EmployeeHTTPEndpoint{
		employeeSubmitAttendanceUsecase,
		employeeSubmitOvertimeUsecase,
		employeeSubmitReimbursementUsecase,
		employeeGeneratePayslipUsecase,
		logger,
		validator,
	}
}

// @Summary      Submit attendance
// @Description  Submit daily attendance record. Employee ID is obtained from JWT token.
// @Tags         Employee
// @Accept       json
// @Produce      json
// @Param        request body usecases.EmployeeSubmitAttendanceInput true "Attendance details (attendance period ID and date only)"
// @Success      200  {object}  usecases.EmployeeSubmitAttendanceOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /employee/attendance [post]
func (ehttp *EmployeeHTTPEndpoint) SubmitAttendance(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.EmployeeSubmitAttendanceInput
	if err := request.Decode(&input); err != nil {
		ehttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ehttp.validator.Struct(input); err != nil {
		ehttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ehttp.employeeSubmitAttendanceUsecase.Execute(ctx, input)
	if err != nil {
		ehttp.logger.Errorw("failed to execute employee submit attendance usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// @Summary      Submit overtime
// @Description  Submit overtime request. Employee ID is obtained from JWT token. Requires attendance period ID to validate the overtime date is within the period range.
// @Tags         Employee
// @Accept       json
// @Produce      json
// @Param        request body usecases.EmployeeSubmitOvertimeInput true "Overtime details (attendance period ID, date, and hours)"
// @Success      200  {object}  usecases.EmployeeSubmitOvertimeOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /employee/overtime [post]
func (ehttp *EmployeeHTTPEndpoint) Overtime(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.EmployeeSubmitOvertimeInput
	if err := request.Decode(&input); err != nil {
		ehttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ehttp.validator.Struct(input); err != nil {
		ehttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ehttp.employeeSubmitOvertimeUsecase.Execute(ctx, input)
	if err != nil {
		ehttp.logger.Errorw("failed to execute employee submit overtime usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// @Summary      Submit reimbursement
// @Description  Submit reimbursement request. Employee ID is obtained from JWT token. Attendance period ID is required.
// @Tags         Employee
// @Accept       json
// @Produce      json
// @Param        request body usecases.EmployeeSubmitReimbursementInput true "Reimbursement details (attendance period ID, amount, and description)"
// @Success      200  {object}  usecases.EmployeeSubmitReimbursementOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /employee/reimbursement [post]
func (ehttp *EmployeeHTTPEndpoint) Reimbursement(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.EmployeeSubmitReimbursementInput
	if err := request.Decode(&input); err != nil {
		ehttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if input.AttendancePeriodID == uuid.Nil {
		ehttp.logger.Errorw("missing attendance period ID")
		return nil, pkgerror.NewBusinessError("attendance period ID is required")
	}
	if input.Amount.Cmp(decimal.Zero) <= 0 {
		ehttp.logger.Errorw("invalid reimbursement amount", "amount", input.Amount)
		return nil, pkgerror.NewBusinessError("amount must be greater than 0")
	}
	if input.Description == "" {
		ehttp.logger.Errorw("missing reimbursement description")
		return nil, pkgerror.NewBusinessError("description is required")
	}

	output, err := ehttp.employeeSubmitReimbursementUsecase.Execute(ctx, input)
	if err != nil {
		ehttp.logger.Errorw("failed to execute employee submit reimbursement usecase", "error", err)
		return nil, err
	}

	return output, nil
}

// @Summary      Generate payslip
// @Description  Generate payslip for a specific month
// @Tags         Employee
// @Accept       json
// @Produce      json
// @Param        request body usecases.EmployeeGeneratePayslipInput true "Payslip generation details"
// @Success      200  {object}  usecases.EmployeeGeneratePayslipOutput
// @Failure      400  {object}  pkgerror.Error
// @Failure      401  {object}  pkgerror.Error
// @Failure      403  {object}  pkgerror.Error
// @Security     BearerAuth
// @Router       /employee/generate-payslip [get]
func (ehttp *EmployeeHTTPEndpoint) GeneratePayslip(ctx context.Context, request pkghttp.Request) (any, error) {
	var input usecases.EmployeeGeneratePayslipInput
	if err := request.Decode(&input); err != nil {
		ehttp.logger.Errorw("failed to decode request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	if err := ehttp.validator.Struct(input); err != nil {
		ehttp.logger.Errorw("failed to validate request body", "error", err)
		return nil, pkgerror.NewBusinessError("invalid request body")
	}

	output, err := ehttp.employeeGeneratePayslipUsecase.Execute(ctx, input)
	if err != nil {
		ehttp.logger.Errorw("failed to execute employee generate payslip usecase", "error", err)
		return nil, err
	}

	return output, nil
}
