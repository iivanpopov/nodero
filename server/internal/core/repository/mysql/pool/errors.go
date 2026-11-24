package pool

import "errors"

var (
	ErrNoRows             = errors.New("no rows in result set")
	ErrDuplicateEntry     = errors.New("duplicate entry")
	ErrViolatesForeignKey = errors.New("violates foreign key constraint")
)
