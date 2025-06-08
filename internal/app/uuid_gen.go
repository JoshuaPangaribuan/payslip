package app

import (
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkguid"
)

func (app *App) initUUIDGen() {
	app.uuidGen = pkguid.NewUUID()
}
