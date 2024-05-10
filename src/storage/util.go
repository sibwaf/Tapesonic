package storage

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var voidLog = logger.Default.LogMode(logger.Silent)

type DbHelper struct {
	*gorm.DB
}

func NewDbHelper(db *gorm.DB) *DbHelper {
	return &DbHelper{db}
}

func (db *DbHelper) ExclusiveTransaction(tx func(*gorm.DB) error) error {
	return db.Connection(func(exclusiveTx *gorm.DB) error {
		exclusiveTxNoLog := exclusiveTx.Session(&gorm.Session{Logger: voidLog})
		exclusiveTxNoTransaction := exclusiveTx.Session(&gorm.Session{SkipDefaultTransaction: true})

		err := exclusiveTxNoLog.Exec("BEGIN EXCLUSIVE TRANSACTION").Error
		if err != nil {
			return err
		}

		err = tx(exclusiveTxNoTransaction)

		if err != nil {
			rollbackErr := exclusiveTxNoLog.Exec("ROLLBACK").Error
			return errors.Join(err, rollbackErr)
		} else {
			return exclusiveTxNoLog.Exec("COMMIT").Error
		}
	})
}
