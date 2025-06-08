package app

import (
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

func (app *App) initLogger() {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		app.err = multierror.Append(app.err, err)
	}

	app.logger = logger
}
