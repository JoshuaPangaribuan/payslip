-- +goose Up
-- +goose StatementBegin
-- 1. Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Enum types
CREATE TYPE employee_role AS ENUM ('employee','admin');
CREATE TYPE audit_action AS ENUM ('INSERT','UPDATE','DELETE');

-- 3. Employee
CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salary_monthly NUMERIC(12,2) NOT NULL CHECK (salary_monthly >= 0),
    role employee_role NOT NULL DEFAULT 'employee',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 4. Attendance Period
CREATE TABLE attendance_period (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL CHECK (end_date >= start_date),
    payroll_run BOOLEAN NOT NULL DEFAULT FALSE,
    payroll_run_id UUID, -- FK: payroll_run.id (relasi 1:1 opsional, penanda)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 5. Attendance Record
CREATE TABLE attendance_record (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL,           -- FK: employee.id
    attendance_period_id UUID NOT NULL,  -- FK: attendance_period.id
    attendance_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 6. Overtime Request
CREATE TABLE overtime_request (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL,           -- FK: employee.id
    attendance_period_id UUID NOT NULL,  -- FK: attendance_period.id
    overtime_date DATE NOT NULL,
    hours INT NOT NULL CHECK (hours BETWEEN 0 AND 3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 7. Reimbursement Request
CREATE TABLE reimbursement_request (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL,           -- FK: employee.id
    attendance_period_id UUID NOT NULL,  -- FK: attendance_period.id
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 8. Payroll Run
CREATE TABLE payroll_run (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendance_period_id UUID NOT NULL UNIQUE, -- FK: attendance_period.id (1:1)
    run_at TIMESTAMPTZ NOT NULL,
    run_by UUID NOT NULL,                      -- FK: employee.id (admin yang eksekusi)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 9. Payslip
CREATE TABLE payslip (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL,       -- FK: employee.id
    payroll_run_id UUID NOT NULL,    -- FK: payroll_run.id
    attendance_pay NUMERIC(12,2) NOT NULL CHECK (attendance_pay >= 0),
    overtime_pay NUMERIC(12,2) NOT NULL CHECK (overtime_pay >= 0),
    reimbursement_total NUMERIC(12,2) NOT NULL CHECK (reimbursement_total >= 0),
    take_home_pay NUMERIC(12,2) NOT NULL CHECK (take_home_pay >= 0),
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    updated_by UUID,     -- FK: employee.id (audit trail)
    ip_address INET,
    request_id UUID
);

-- 10. Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(255) NOT NULL,
    record_id UUID NOT NULL,
    action audit_action NOT NULL,
    changes VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,     -- FK: employee.id (audit trail)
    ip_address VARCHAR(255),
    request_id UUID
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS payslip CASCADE;
DROP TABLE IF EXISTS payroll_run CASCADE;
DROP TABLE IF EXISTS reimbursement_request CASCADE;
DROP TABLE IF EXISTS overtime_request CASCADE;
DROP TABLE IF EXISTS attendance_record CASCADE;
DROP TABLE IF EXISTS attendance_period CASCADE;
DROP TABLE IF EXISTS employee CASCADE;

DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS employee_role;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
