package service

import "errors"

var (
	ErrEventNotFound = errors.New("EVENT_NOT_FOUND")
	ErrAccessDenied  = errors.New("ACCESS_DENIED")
)
