package app

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

func (app *App) initValidator() {
	app.validator = validator.New()

	app.validator.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		val, ok := field.Interface().(decimal.Decimal)
		if ok {
			return val.IntPart()
		}

		return nil
	}, decimal.Decimal{})
}
