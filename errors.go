package token

import "errors"

var (
	ErrTokenNotFound  = errors.New(`token not found`)
	ErrTokenIsEmpty   = errors.New(`parameter "token" is empty`)
	ErrTokenIsExpired = errors.New(`token is expired`)
)
