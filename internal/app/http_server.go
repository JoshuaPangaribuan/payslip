package app

import (
	"net/http"
	"time"
)

func (app *App) makeHTTPServer() {
	addr := app.config.GetString("server.address.http")

	httpServer := &http.Server{
		Addr:    addr,
		Handler: app.router,

		ReadHeaderTimeout: 1 * time.Second,
	}

	app.httpServer = httpServer
}
