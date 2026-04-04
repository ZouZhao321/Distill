package domain

import "errors"

// 预定义的业务错误。
var (
	ErrAlreadyExists = errors.New("asset already exists")
	ErrNotFound      = errors.New("asset not found")
	ErrInvalidPath   = errors.New("invalid path")
	ErrStoreClosed   = errors.New("store is closed")
	ErrHashMismatch  = errors.New("hash mismatch")
	ErrEmptySource   = errors.New("source is empty")
	ErrNotDirectory  = errors.New("path is not a directory")
)
