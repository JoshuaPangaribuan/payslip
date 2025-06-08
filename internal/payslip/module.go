package payslip

import (
	"database/sql"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/gateways"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/interactors"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgsql"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkguid"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Exposed struct{}

type PayslipDependencies struct {
	DB           *sql.DB
	Logger       *zap.SugaredLogger
	QueryBuilder pkgsql.GoquBuilder
	UUIDV4Gen    pkguid.UUID
	Router       *httprouter.Router
	Validator    *validator.Validate
	Config       *viper.Viper
}

func NewPayslipDomain(deps PayslipDependencies) *Exposed {
	payslipGateway := gateways.NewPayslipSQLGateway(deps.DB, deps.Logger, deps.QueryBuilder)

	// Login
	employeeLoginUsecase := interactors.NewEmployeeLoginInteractor(payslipGateway, deps.Config.GetString("jwt.secret"), deps.Logger)

	commonHTTPEndpoint := gateways.NewCommonHTTPEndpoint(
		employeeLoginUsecase,
		deps.Logger,
		deps.Validator,
	)

	// Admin
	adminAddAttendancePeriodUsecase := interactors.NewAdminAddAttendancePeriodInteractor(payslipGateway, payslipGateway, deps.UUIDV4Gen, deps.Logger)
	adminRunPayrollUsecase := interactors.NewAdminRunPayrollInteractor(payslipGateway, payslipGateway, deps.UUIDV4Gen)
	adminPayrollSummaryUsecase := interactors.NewAdminPayrollSummaryInteractor(payslipGateway)

	adminHTTPEndpoint := gateways.NewAdminHTTPEndpoint(
		adminAddAttendancePeriodUsecase,
		adminRunPayrollUsecase,
		adminPayrollSummaryUsecase,
		deps.Logger,
		deps.Validator,
	)

	// Employee
	employeeSubmitAttendanceUsecase := interactors.NewEmployeeSubmitAttendanceInteractor(payslipGateway, payslipGateway, deps.UUIDV4Gen, deps.Logger)
	employeeSubmitOvertimeUsecase := interactors.NewEmployeeSubmitOvertimeInteractor(payslipGateway, payslipGateway, deps.UUIDV4Gen, deps.Logger, payslipGateway)
	employeeSubmitReimbursementUsecase := interactors.NewEmployeeSubmitReimbursementInteractor(payslipGateway, payslipGateway, deps.Logger, deps.UUIDV4Gen)
	employeeGeneratePayslipUsecase := interactors.NewEmployeeGeneratePayslipInteractor(payslipGateway, deps.UUIDV4Gen)

	employeeHTTPEndpoint := gateways.NewEmployeeHTTPEndpoint(
		employeeSubmitAttendanceUsecase,
		employeeSubmitOvertimeUsecase,
		employeeSubmitReimbursementUsecase,
		employeeGeneratePayslipUsecase,
		deps.Logger,
		deps.Validator,
	)

	gateways.NewPayslipHttpGateway(
		deps.Router,
		deps.Logger,
		deps.Validator,
		adminHTTPEndpoint,
		employeeHTTPEndpoint,
		commonHTTPEndpoint,
	)

	return &Exposed{}
}
