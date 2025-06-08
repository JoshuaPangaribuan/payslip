package app

import (
	"database/sql"
	"fmt"
	"time"

	// _ "github.com/go-sql-driver/mysql" //nolint:gosec // won't be an issue
	_ "github.com/lib/pq" //nolint:gosec // won't be an issue

	"github.com/doug-martin/goqu/v9"
	"github.com/hashicorp/go-multierror"
)

func (app *App) setUpGoqu() {
	loc, _ := time.LoadLocation("Asia/Jakarta") //nolint:errcheck // won't be an error

	goqu.SetTimeLocation(loc)

	dialect := app.config.GetString("database.query.dialect")

	app.queryBuilder = goqu.New(dialect, app.database)
}

type dbConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
}

func (app *App) initDB() {
	host := app.config.GetString("database.host")
	port := app.config.GetString("database.port")
	user := app.config.GetString("database.user")
	password := app.config.GetString("database.password")
	databaseName := app.config.GetString("database.name")

	dbConfig := dbConfig{
		Host:         host,
		Port:         port,
		User:         user,
		Password:     password,
		DatabaseName: databaseName,
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DatabaseName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		app.err = multierror.Append(app.err, err)
	}

	if err := db.Ping(); err != nil {
		app.err = multierror.Append(app.err, err)
	}

	app.database = db
}
