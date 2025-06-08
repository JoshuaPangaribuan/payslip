package app

import "context"

func (app *App) setUpClosers() {
	app.closersFn = append(app.closersFn, []func(context.Context) error{
		func(ctx context.Context) error {
			return app.httpServer.Shutdown(ctx)
		},
	}...)
}
