package main

import (
	_ "github.com/JoshuaPangaribuan/payslip/docs"
	"github.com/JoshuaPangaribuan/payslip/internal/app"
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

// @host localhost:8081
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	app.Run()
}
