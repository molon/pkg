package orm

import (
	"github.com/jinzhu/gorm"
)

func RunTx(db *gorm.DB, handler func(tx *gorm.DB) error) (err error) {
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r)
			} else {
				tx.Commit()
			}

		}
	}()
	err = handler(tx)
	return
}
