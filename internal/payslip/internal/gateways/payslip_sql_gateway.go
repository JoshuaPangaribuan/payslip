package gateways

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/interactors"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgsql"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _ interactors.AttendancePeriodRepository = (*PayslipSQLGateway)(nil)
var _ interactors.PayrollSummaryRepository = (*PayslipSQLGateway)(nil)
var _ interactors.PayrollRepository = (*PayslipSQLGateway)(nil)
var _ interactors.EmployeeRepository = (*PayslipSQLGateway)(nil)
var _ interactors.AttendanceRepository = (*PayslipSQLGateway)(nil)
var _ interactors.OvertimeRepository = (*PayslipSQLGateway)(nil)
var _ interactors.ReimbursementRepository = (*PayslipSQLGateway)(nil)
var _ interactors.PayslipRepository = (*PayslipSQLGateway)(nil)
var _ interactors.AuditLogRepository = (*PayslipSQLGateway)(nil)

type PayslipSQLGateway struct {
	db           pkgsql.SQL
	logger       *zap.SugaredLogger
	queryBuilder pkgsql.GoquBuilder

	employeeTableName             string
	attendancePeriodTableName     string
	attendanceRecordTableName     string
	overtimeRequestTableName      string
	reimbursementRequestTableName string
	payrollRunTableName           string
	payslipTableName              string
	auditLogTableName             string
}

func NewPayslipSQLGateway(
	db *sql.DB,
	logger *zap.SugaredLogger,
	queryBuilder pkgsql.GoquBuilder,
) *PayslipSQLGateway {
	return &PayslipSQLGateway{
		db:           db,
		logger:       logger,
		queryBuilder: queryBuilder,

		employeeTableName:             "employee",
		attendancePeriodTableName:     "attendance_period",
		attendanceRecordTableName:     "attendance_record",
		overtimeRequestTableName:      "overtime_request",
		reimbursementRequestTableName: "reimbursement_request",
		payrollRunTableName:           "payroll_run",
		payslipTableName:              "payslip",
		auditLogTableName:             "audit_log",
	}
}

// CreateAttendancePeriod creates a new attendance period
func (g *PayslipSQLGateway) CreateAttendancePeriod(ctx context.Context, period *sqlentity.AttendancePeriods) error {
	cols := make([]interface{}, len(period.StringColumns()))
	for i, col := range period.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(period.DriverValues()))
	for i, val := range period.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.attendancePeriodTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create attendance period query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create attendance period", "error", err)
		return err
	}

	return nil
}

// GetActiveEmployees gets all active employees
func (g *PayslipSQLGateway) GetActiveEmployees(ctx context.Context) ([]*sqlentity.Employees, error) {
	query := g.queryBuilder.From(g.employeeTableName).
		Select("*").
		Where(goqu.C("status").Eq("active"))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get active employees query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get active employees", "error", err)
		return nil, err
	}
	defer rows.Close()

	var employees []*sqlentity.Employees
	for rows.Next() {
		var employee sqlentity.Employees
		err := rows.Scan(employee.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan employee", "error", err)
			return nil, err
		}
		employees = append(employees, &employee)
	}

	return employees, nil
}

// CreatePayrollRun creates a new payroll run
func (g *PayslipSQLGateway) CreatePayrollRun(ctx context.Context, run *sqlentity.PayrollRuns) error {
	cols := make([]interface{}, len(run.StringColumns()))
	for i, col := range run.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(run.DriverValues()))
	for i, val := range run.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.payrollRunTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create payroll run query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create payroll run", "error", err)
		return err
	}

	return nil
}

// GetByUsername gets employee by username
func (g *PayslipSQLGateway) GetByUsername(ctx context.Context, username string) (*sqlentity.Employees, error) {
	query := g.queryBuilder.From(g.employeeTableName).
		Select("*").
		Where(goqu.C("username").Eq(username))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get employee by username query", "error", err)
		return nil, err
	}

	var employee sqlentity.Employees
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(employee.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get employee by username", "error", err)
		return nil, err
	}

	return &employee, nil
}

// CreateAttendance creates a new attendance record
func (g *PayslipSQLGateway) CreateAttendance(ctx context.Context, attendance *sqlentity.AttendanceRecords) error {
	cols := make([]interface{}, len(attendance.StringColumns()))
	for i, col := range attendance.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(attendance.DriverValues()))
	for i, val := range attendance.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.attendanceRecordTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create attendance query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create attendance", "error", err)
		return err
	}

	return nil
}

// CreateOvertime creates a new overtime request
func (g *PayslipSQLGateway) CreateOvertime(ctx context.Context, overtime *sqlentity.OvertimeRequests) error {
	g.logger.Infow("starting overtime creation",
		"overtime_id", overtime.ID,
		"employee_id", overtime.EmployeeID,
		"attendance_period_id", overtime.AttendancePeriodID,
		"date", overtime.OvertimeDate,
		"hours", overtime.Hours,
	)

	cols := make([]interface{}, len(overtime.StringColumns()))
	for i, col := range overtime.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(overtime.DriverValues()))
	for i, val := range overtime.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.overtimeRequestTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create overtime query",
			"error", err,
			"overtime_id", overtime.ID,
			"employee_id", overtime.EmployeeID,
			"table", g.overtimeRequestTableName,
		)
		return err
	}

	g.logger.Infow("executing overtime creation query",
		"overtime_id", overtime.ID,
		"employee_id", overtime.EmployeeID,
		"sql", sqlStr,
		"args", args,
	)

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create overtime",
			"error", err,
			"overtime_id", overtime.ID,
			"employee_id", overtime.EmployeeID,
			"attendance_period_id", overtime.AttendancePeriodID,
			"date", overtime.OvertimeDate,
			"hours", overtime.Hours,
			"sql", sqlStr,
			"args", args,
		)
		return err
	}

	g.logger.Infow("successfully created overtime",
		"overtime_id", overtime.ID,
		"employee_id", overtime.EmployeeID,
		"attendance_period_id", overtime.AttendancePeriodID,
		"date", overtime.OvertimeDate,
		"hours", overtime.Hours,
	)

	return nil
}

// CreateReimbursement creates a new reimbursement request
func (g *PayslipSQLGateway) CreateReimbursement(ctx context.Context, reimbursement *sqlentity.ReimbursementRequests) error {
	cols := make([]interface{}, len(reimbursement.StringColumns()))
	for i, col := range reimbursement.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(reimbursement.DriverValues()))
	for i, val := range reimbursement.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.reimbursementRequestTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create reimbursement query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create reimbursement", "error", err)
		return err
	}

	return nil
}

// GetByID gets attendance period by ID
func (g *PayslipSQLGateway) GetByID(ctx context.Context, id uuid.UUID) (*sqlentity.AttendancePeriods, error) {
	query := g.queryBuilder.From(g.attendancePeriodTableName).
		Select("*").
		Where(goqu.C("id").Eq(id))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance period by ID query", "error", err)
		return nil, err
	}

	var period sqlentity.AttendancePeriods
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(period.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get attendance period by ID", "error", err)
		return nil, err
	}

	return &period, nil
}

// GetAttendanceByPeriod gets attendance records by period
func (g *PayslipSQLGateway) GetAttendanceByPeriod(ctx context.Context, periodID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.AttendanceRecords, error) {
	query := g.queryBuilder.From(g.attendanceRecordTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("period_id").Eq(periodID),
			goqu.C("attendance_date").Gte(startDate),
			goqu.C("attendance_date").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance by period query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get attendance by period", "error", err)
		return nil, err
	}
	defer rows.Close()

	var records []*sqlentity.AttendanceRecords
	for rows.Next() {
		var record sqlentity.AttendanceRecords
		err := rows.Scan(record.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan attendance record", "error", err)
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetAttendanceByEmployeeAndDate gets attendance record by employee ID and date
func (g *PayslipSQLGateway) GetAttendanceByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*sqlentity.AttendanceRecords, error) {
	query := g.queryBuilder.From(g.attendanceRecordTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.C("attendance_date").Eq(date),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance by employee and date query", "error", err)
		return nil, err
	}

	var record sqlentity.AttendanceRecords
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(record.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get attendance by employee and date", "error", err)
		return nil, err
	}

	return &record, nil
}

// GetOvertimeByEmployeeAndDate gets overtime request by employee ID and date
func (g *PayslipSQLGateway) GetOvertimeByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*sqlentity.OvertimeRequests, error) {
	query := g.queryBuilder.From(g.overtimeRequestTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.C("overtime_date").Eq(date),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get overtime by employee and date query", "error", err)
		return nil, err
	}

	var request sqlentity.OvertimeRequests
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(request.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get overtime by employee and date", "error", err)
		return nil, err
	}

	return &request, nil
}

// GetAttendancePeriodByDateRange gets attendance period by date range
func (g *PayslipSQLGateway) GetAttendancePeriodByDateRange(ctx context.Context, startDate, endDate time.Time) (*sqlentity.AttendancePeriods, error) {
	query := g.queryBuilder.From(g.attendancePeriodTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("start_date").Lte(endDate),
			goqu.C("end_date").Gte(startDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance period query", "error", err)
		return nil, err
	}

	var period sqlentity.AttendancePeriods
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(period.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get attendance period", "error", err)
		return nil, err
	}

	return &period, nil
}

// GetAttendancePeriod gets attendance period by ID
func (g *PayslipSQLGateway) GetAttendancePeriod(ctx context.Context, id uuid.UUID) (*sqlentity.AttendancePeriods, error) {
	return g.GetByID(ctx, id)
}

// GetOvertimeByPeriod gets overtime requests by period
func (g *PayslipSQLGateway) GetOvertimeByPeriod(ctx context.Context, periodID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.OvertimeRequests, error) {
	query := g.queryBuilder.From(g.overtimeRequestTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("period_id").Eq(periodID),
			goqu.C("overtime_date").Gte(startDate),
			goqu.C("overtime_date").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get overtime by period query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get overtime by period", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*sqlentity.OvertimeRequests
	for rows.Next() {
		var request sqlentity.OvertimeRequests
		err := rows.Scan(request.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan overtime request", "error", err)
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

// GetPayrollRun gets payroll run by ID
func (g *PayslipSQLGateway) GetPayrollRun(ctx context.Context, id uuid.UUID) (*sqlentity.PayrollRuns, error) {
	query := g.queryBuilder.From(g.payrollRunTableName).
		Select("*").
		Where(goqu.C("id").Eq(id))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get payroll run query", "error", err)
		return nil, err
	}

	var run sqlentity.PayrollRuns
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(run.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get payroll run", "error", err)
		return nil, err
	}

	return &run, nil
}

// GetPayrollRunByAttendancePeriod gets payroll run by attendance period ID
func (g *PayslipSQLGateway) GetPayrollRunByAttendancePeriod(ctx context.Context, periodID uuid.UUID) (*sqlentity.PayrollRuns, error) {
	query := g.queryBuilder.From(g.payrollRunTableName).
		Select("*").
		Where(goqu.C("attendance_period_id").Eq(periodID))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get payroll run by attendance period query", "error", err)
		return nil, err
	}

	var run sqlentity.PayrollRuns
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(run.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get payroll run by attendance period", "error", err)
		return nil, err
	}

	return &run, nil
}

// GetReimbursementsByPeriod gets reimbursement requests by period
func (g *PayslipSQLGateway) GetReimbursementsByPeriod(ctx context.Context, periodID uuid.UUID, startDate, endDate time.Time) ([]*sqlentity.ReimbursementRequests, error) {
	query := g.queryBuilder.From(g.reimbursementRequestTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("period_id").Eq(periodID),
			goqu.C("created_at").Gte(startDate),
			goqu.C("created_at").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get reimbursements by period query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get reimbursements by period", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*sqlentity.ReimbursementRequests
	for rows.Next() {
		var request sqlentity.ReimbursementRequests
		err := rows.Scan(request.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan reimbursement request", "error", err)
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

// CreatePayslip creates a new payslip
func (g *PayslipSQLGateway) CreatePayslip(ctx context.Context, payslip *sqlentity.Payslips) error {
	cols := make([]interface{}, len(payslip.StringColumns()))
	for i, col := range payslip.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(payslip.DriverValues()))
	for i, val := range payslip.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.payslipTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create payslip query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create payslip", "error", err)
		return err
	}

	return nil
}

// GetAttendanceRecordsByEmployeeAndDateRange gets attendance records by employee ID and date range
func (g *PayslipSQLGateway) GetAttendanceRecordsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.AttendanceRecords, error) {
	query := g.queryBuilder.From(g.attendanceRecordTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.C("attendance_date").Gte(startDate),
			goqu.C("attendance_date").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance records query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get attendance records by employee and date range", "error", err)
		return nil, err
	}
	defer rows.Close()

	var records []*sqlentity.AttendanceRecords
	for rows.Next() {
		var record sqlentity.AttendanceRecords
		err := rows.Scan(record.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan attendance record", "error", err)
			return nil, err
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetEmployeeByID gets employee by ID
func (g *PayslipSQLGateway) GetEmployeeByID(ctx context.Context, id string) (*sqlentity.Employees, error) {
	query := g.queryBuilder.From(g.employeeTableName).
		Select("*").
		Where(goqu.C("id").Eq(id))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get employee query", "error", err)
		return nil, err
	}

	var employee sqlentity.Employees
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(employee.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get employee by ID", "error", err)
		return nil, err
	}

	return &employee, nil
}

// GetOvertimeRequestsByEmployeeAndDateRange gets overtime requests by employee ID and date range
func (g *PayslipSQLGateway) GetOvertimeRequestsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.OvertimeRequests, error) {
	query := g.queryBuilder.From(g.overtimeRequestTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.C("overtime_date").Gte(startDate),
			goqu.C("overtime_date").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get overtime requests query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get overtime requests by employee and date range", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*sqlentity.OvertimeRequests
	for rows.Next() {
		var request sqlentity.OvertimeRequests
		err := rows.Scan(request.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan overtime request", "error", err)
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

// GetPayrollRunByMonthAndYear gets payroll run by month and year
func (g *PayslipSQLGateway) GetPayrollRunByMonthAndYear(ctx context.Context, month, year int) (*sqlentity.PayrollRuns, error) {
	query := g.queryBuilder.From(g.payrollRunTableName).
		Select("*").
		Where(goqu.And(
			goqu.Func("EXTRACT", goqu.L("MONTH FROM run_at")).Eq(month),
			goqu.Func("EXTRACT", goqu.L("YEAR FROM run_at")).Eq(year),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get payroll run query", "error", err)
		return nil, err
	}

	var run sqlentity.PayrollRuns
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(run.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get payroll run by month and year", "error", err)
		return nil, err
	}

	return &run, nil
}

// GetPayslipByEmployeeAndPeriod gets payslip by employee ID and period
func (g *PayslipSQLGateway) GetPayslipByEmployeeAndPeriod(ctx context.Context, employeeID string, month, year int) (*sqlentity.Payslips, error) {
	query := g.queryBuilder.From(g.payslipTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.Func("EXTRACT", goqu.L("MONTH FROM generated_at")).Eq(month),
			goqu.Func("EXTRACT", goqu.L("YEAR FROM generated_at")).Eq(year),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get payslip query", "error", err)
		return nil, err
	}

	var payslip sqlentity.Payslips
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(payslip.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get payslip by employee and period", "error", err)
		return nil, err
	}

	return &payslip, nil
}

// GetReimbursementRequestsByEmployeeAndDateRange gets reimbursement requests by employee ID and date range
func (g *PayslipSQLGateway) GetReimbursementRequestsByEmployeeAndDateRange(ctx context.Context, employeeID string, startDate, endDate time.Time) ([]*sqlentity.ReimbursementRequests, error) {
	query := g.queryBuilder.From(g.reimbursementRequestTableName).
		Select("*").
		Where(goqu.And(
			goqu.C("employee_id").Eq(employeeID),
			goqu.C("created_at").Gte(startDate),
			goqu.C("created_at").Lte(endDate),
		))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get reimbursement requests query", "error", err)
		return nil, err
	}

	rows, err := g.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to get reimbursement requests by employee and date range", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*sqlentity.ReimbursementRequests
	for rows.Next() {
		var request sqlentity.ReimbursementRequests
		err := rows.Scan(request.Values()...)
		if err != nil {
			g.logger.Errorw("failed to scan reimbursement request", "error", err)
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

// CreateAuditLog creates a new audit log entry
func (g *PayslipSQLGateway) CreateAuditLog(ctx context.Context, log *sqlentity.AuditLogs) error {
	cols := make([]interface{}, len(log.StringColumns()))
	for i, col := range log.StringColumns() {
		cols[i] = col
	}

	vals := make([]interface{}, len(log.DriverValues()))
	for i, val := range log.DriverValues() {
		vals[i] = val
	}

	query := g.queryBuilder.Insert(g.auditLogTableName).
		Cols(cols...).
		Vals(vals)

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build create audit log query", "error", err)
		return err
	}

	_, err = g.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		g.logger.Errorw("failed to create audit log", "error", err)
		return err
	}

	return nil
}

// GetAttendancePeriodByID gets attendance period by ID
func (g *PayslipSQLGateway) GetAttendancePeriodByID(ctx context.Context, id uuid.UUID) (*sqlentity.AttendancePeriods, error) {
	query := g.queryBuilder.From(g.attendancePeriodTableName).
		Select("*").
		Where(goqu.C("id").Eq(id))

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		g.logger.Errorw("failed to build get attendance period by ID query", "error", err)
		return nil, err
	}

	var period sqlentity.AttendancePeriods
	err = g.db.QueryRowContext(ctx, sqlStr, args...).Scan(period.Values()...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		g.logger.Errorw("failed to get attendance period by ID", "error", err)
		return nil, err
	}

	return &period, nil
}
