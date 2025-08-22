package service

import "errors"

var (
	// ErrRequiredArgumentIsMissing returned when user request is missing a required argument.
	ErrRequiredArgumentIsMissing = errors.New("one of required arguments is missing")
)
