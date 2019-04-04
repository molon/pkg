package orm

import (
	"context"

	"github.com/jinzhu/gorm"
)

const inTxKey = "gorm.inTx"

func RunTx(ctx context.Context, db *gorm.DB,
	handler func(ctx context.Context, db *gorm.DB) error,
) (err error) {
	// 已经在事务中，就直接执行即可
	if _, ok := ctx.Value(inTxKey).(bool); ok {
		return handler(ctx, db)
	}

	// 开启事务执行
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

	ctx = context.WithValue(ctx, inTxKey, true)
	err = handler(ctx, tx)
	return
}
