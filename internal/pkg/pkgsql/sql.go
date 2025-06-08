package pkgsql

import (
	"github.com/doug-martin/goqu/v9"
)

type SQL interface {
	goqu.SQLDatabase
}

type GoquBuilder interface {
	From(from ...interface{}) *goqu.SelectDataset
	Insert(table interface{}) *goqu.InsertDataset
	Select(cols ...interface{}) *goqu.SelectDataset
	Update(table interface{}) *goqu.UpdateDataset
}
