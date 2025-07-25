package domain

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	//ErrDBQuery        = errors.New("database query error")
	//ErrDuplicateEntry = errors.New("duplicate entry")
	//ErrTimeout        = errors.New("database operation timeout")
)
