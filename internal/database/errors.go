package database

import "errors"

// Sentinel errors returned by Store implementations and the
// decorator wrappers. Handlers map these to HTTP status codes.
var (
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
)
