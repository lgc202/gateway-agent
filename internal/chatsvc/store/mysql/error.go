package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"

	mysqldriver "github.com/go-sql-driver/mysql"

	"github.com/lgc202/gateway-agent/internal/pkg/errorsx"
)

// databaseError 将数据库错误转换为稳定的应用错误，并保留底层原因供服务端日志使用。
func databaseError(err error) error {
	if isTransientDatabaseError(err) {
		return errorsx.Wrap(errorsx.CodeDependencyUnavailable, "database unavailable", err)
	}

	return errorsx.Wrap(errorsx.CodeInternal, "database operation failed", err)
}

func isTransientDatabaseError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, sql.ErrConnDone) ||
		errors.Is(err, mysqldriver.ErrInvalidConn) {
		return true
	}

	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}

	mysqlError, ok := errors.AsType[*mysqldriver.MySQLError](err)
	if !ok {
		return false
	}

	switch mysqlError.Number {
	case 1040,
		1158, 1159, 1160, 1161,
		1205, 1213,
		2002, 2003, 2006, 2013:
		return true
	default:
		return false
	}
}
