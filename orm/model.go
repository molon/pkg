package orm

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
)

type Model struct {
	Id        string    `gorm:"size:20;primary_key"`
	CreatedAt time.Time `sql:"index"`
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Version   int64
}

func (m *Model) BeforeCreate(scope *gorm.Scope) error {
	if len(m.Id) > 0 {
		return nil
	}

	return scope.SetColumn("Id", xid.New().String())
}
