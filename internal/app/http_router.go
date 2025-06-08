package app

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *App) initRouter() {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode("not found")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := json.NewEncoder(w).Encode("method not allowed")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	app.router = router
}
