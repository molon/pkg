package orm

import (
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

func IsDuplicateError(err error) bool {
	switch oe := err.(type) {
	case (*pq.Error):
		if oe.Code == "23505" {
			return true
		}
	case (*mysql.MySQLError):
		if oe.Number == 1062 {
			return true
		}
	}

	return false
}
