# Payslip API: Employee Payroll & Attendance System

## What is the Payslip API?

The Payslip API is a system that helps companies manage employee attendance, overtime, reimbursements, and payroll (salary) calculations automatically. It makes sure that:
- Employees can submit their daily attendance, overtime, and reimbursement requests online.
- Admins can set up attendance periods, run payroll, and see salary summaries for all employees.
- Every action is tracked for transparency and auditing (who did what, when, and from where).

**In short:** It automates and tracks everything needed for monthly salary calculation, including extra work and expenses, with full traceability.

---

## What Can You Do With Payslip API?

- **Employees:**
  - Submit daily attendance (only on weekdays).
  - Submit overtime requests (up to 3 hours per day, any day).
  - Submit reimbursement requests (for work-related expenses).
  - Generate and view their payslip (salary breakdown for the month).

- **Admins:**
  - Add new attendance periods (define which days count for payroll).
  - Run payroll for a period (finalize salaries for all employees).
  - View payroll summaries (see total salary payouts and breakdowns).

- **Auditing & Traceability:**
  - Every record (attendance, overtime, reimbursement, payslip) includes who created/updated it, when, from which IP, and with which request ID.
  - All important actions are logged in an audit log for compliance and review.

---

## Project Structure & Design Pattern

The project is organized for clarity and scalability:

- `internal/` : Main application code, split by business logic (usecases), data access (gateways), and business rules (interactors).
- `docs/` : API documentation and static files.
- `migration/` : Database migration scripts (for setting up the database tables).
- `seed/` : Scripts to fill the database with fake employees and admin for testing.
- `Makefile` : Easy-to-use commands for running, testing, and managing the app.
- `Dockerfile` & `docker-compose.yaml` : For running everything in containers (no manual setup needed).

**Pattern Used:**
- The code uses a clean architecture (hexagonal/onion), separating business logic from data access and HTTP/API layers. This makes it easy to test, maintain, and extend.

---

## Understanding the Database (ERD.png)

The file `ERD.png` shows how all the data is connected. Here's what each part means:

- **EMPLOYEE**: Stores employee info (name, salary, role, etc.).
- **ATTENDANCE_PERIOD**: Defines the start and end dates for each payroll period.
- **ATTENDANCE_RECORD**: Each time an employee checks in, a record is created here.
- **OVERTIME_REQUEST**: When employees work extra hours, they submit a request here.
- **REIMBURSEMENT_REQUEST**: Employees can ask for money back for work expenses.
- **PAYROLL_RUN**: When the admin runs payroll, a record is created here for that period.
- **PAYSLIP**: The final salary breakdown for each employee, including attendance, overtime, and reimbursements.
- **AUDIT_LOG**: Tracks every important action for transparency (who did what, when, from where).

**All tables include:**
- When the record was created/updated
- Who did it (user ID)
- From which IP address
- With which request ID (for tracing)

---

## How to Run Payslip API (No Setup Needed!)

**You don't need to install anything except Docker!**

1. **Install Docker** (if you don't have it): [Get Docker here](https://www.docker.com/products/docker-desktop/)
2. **Open a terminal/command prompt in the project folder.**
3. **Use these simple commands to manage everything:**

### Main Docker Commands (from the Makefile)

- `make up-docker`  
  Starts the app and the database in Docker containers. It also runs all the database setup and fills it with fake employees and admin. Use this to start the system.

- `make clean-docker`  
  Stops everything and cleans up the database. Use this if you want to reset or stop the system completely.

- `make restart-docker`  
  Does a full clean and then starts everything again from scratch. Use this if you want to refresh the whole system.

**Tip:**
- These commands do everything for you no manual setup needed!
- Always use the commands ending with `-docker` for the easiest and safest experience.

---

## How to Test the API (No Coding Needed!)

You can try out all the features using Postman (a free tool for API testing):

1. **Install Postman:** [Download here](https://www.postman.com/downloads/)
2. **Open Postman** and import these two files:
   - `Payslip.postman_collection.json` (all the API requests)
   - `Payslip Env.postman_environment.json` (environment variables for easy testing)
3. **Select the imported environment** in Postman (top right dropdown).
4. **Try the requests!**
   - Start with the login requests (admin or employee)
   - Use the other requests to submit attendance, overtime, reimbursement, or run payroll
   - All requests are ready to use just click and go!

**Tip:**
- The collection includes both successful and error cases, so you can see how the system responds to different situations.
- You don't need to write any code or know how APIs work just click the requests in order.

---

## Where to Find More Info

- **API Documentation:**
  - Once the app is running, open [http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html) in your browser for interactive docs.
  - For online, open [https://joshuapangaribuan.github.io/payslip/](https://joshuapangaribuan.github.io/payslip/)
- **ERD Diagram:**
  - See `ERD.png` in the project folder for a visual overview of the data.

---

## Need Help?

If you have any questions or issues, check the documentation or open an issue on the GitHub repository.

---

**Enjoy using Payslip API!**
