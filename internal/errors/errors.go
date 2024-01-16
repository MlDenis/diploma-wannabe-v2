package errors

import "errors"

var ErrValidation error = errors.New("validation error")
var ErrDatabaseUnreachable error = errors.New("database unreachable")
var ErrDatabaseSQLQuery error = errors.New("error with SQL query")
var ErrDatabaseMigration error = errors.New("error with migrations")
var ErrJobChannelClosed error = errors.New("jobs channel closed")
