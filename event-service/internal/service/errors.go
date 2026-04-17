package service

import "errors"

var (
	ErrEventNotFound    = errors.New("EVENT_NOT_FOUND")
	ErrAccessDenied     = errors.New("ACCESS_DENIED")
	ErrAlreadyFavorited = errors.New("ALREADY_FAVORITED")
	ErrFavoriteNotFound = errors.New("FAVORITE_NOT_FOUND")
)
