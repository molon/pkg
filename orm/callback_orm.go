package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
)

type CallbackORM struct{}

func (*CallbackORM) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("Id", xid.New().String())
}
