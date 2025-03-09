package storage

import (
	"errors"
	"fmt"
	"strings"
	"tapesonic/util"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	voidLog = logger.Default.LogMode(logger.Silent)
)

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

func MakeTextSearchCondition(fields []string, query string) string {
	terms := ExtractSearchTerms(query)
	if len(terms) == 0 {
		return ""
	}

	searchField := "''"
	for _, field := range fields {
		searchField = fmt.Sprintf("%s || ' ' || coalesce(%s, '')", searchField, field)
	}

	filter := []string{}
	for _, term := range terms {
		term = EscapeTextLiteralForLike(term, "\\")
		filter = append(filter, fmt.Sprintf("%s LIKE '%% %s%%' ESCAPE '%s'", searchField, term, "\\"))
	}

	return strings.Join(filter, " AND ")
}

func ExtractSearchTerms(query string) []string {
	query, err := util.NormalizeUnicode(query)
	if err != nil {
		return []string{}
	}

	return util.SplitWords(query)
}

func EscapeTextLiteralForLike(str string, escape string) string {
	str = EscapeTextLiteral(str)
	str = strings.ReplaceAll(str, escape, escape+escape)
	str = strings.ReplaceAll(str, "_", escape+"_")
	str = strings.ReplaceAll(str, "%", escape+"%")
	return str
}

func EscapeTextLiteral(str string) string {
	return strings.ReplaceAll(str, "'", "''")
}
