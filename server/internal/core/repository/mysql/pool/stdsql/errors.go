package stdsql

import (
	"database/sql"
	"errors"
	"fmt"

	"server/internal/core/repository/mysql/pool"

	"github.com/go-sql-driver/mysql"
)

const (
	mysqlErrDuplicateEntry = 1062
	mysqlErrForeignKey     = 1452
)

func (p *Pool) MapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %v", pool.ErrNoRows, err)
	}

	if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		switch mysqlErr.Number {
		case mysqlErrDuplicateEntry:
			return fmt.Errorf("%w: %v", pool.ErrDuplicateEntry, err)
		case mysqlErrForeignKey:
			return fmt.Errorf("%w: %v", pool.ErrViolatesForeignKey, err)
		}
	}

	return err
}
