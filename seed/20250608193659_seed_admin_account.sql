-- +goose Up
-- +goose StatementBegin

-- Enable pgcrypto extension for bcrypt
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- First insert admin and get its ID
WITH admin_insert AS (
    INSERT INTO employee (
        username, 
        password_hash, 
        salary_monthly, 
        role,
        created_by,
        updated_by,
        ip_address,
        request_id
    )
    VALUES (
        'admin',
        crypt('admin', gen_salt('bf', 10)),
        0,
        'admin',
        NULL,  -- created_by
        NULL,  -- updated_by
        '127.0.0.1',  -- local IP
        uuid_generate_v4()   -- request_id
    )
    RETURNING id
)
-- Then insert employees using admin's ID
INSERT INTO employee (
    username, 
    password_hash, 
    salary_monthly, 
    role,
    created_by,
    updated_by,
    ip_address,
    request_id
)
SELECT 
    'employee_' || LPAD(generate_series::text, 3, '0') as username,
    crypt('employee_' || LPAD(generate_series::text, 3, '0'), gen_salt('bf', 10)) as password_hash,
    CASE 
        WHEN generate_series % 3 = 0 THEN 20000000 -- Every 3rd employee
        WHEN generate_series % 3 = 1 THEN 15000000 -- Every 1st employee
        ELSE 10000000 -- Every 2nd employee
    END as salary_monthly,
    'employee'::employee_role as role,
    admin_id as created_by,
    admin_id as updated_by,
    '127.0.0.1' as ip_address,
    uuid_generate_v4() as request_id
FROM generate_series(1, 100)
CROSS JOIN (SELECT id as admin_id FROM admin_insert) admin;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM employee WHERE username = 'admin' OR username LIKE 'employee_%';
DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd
