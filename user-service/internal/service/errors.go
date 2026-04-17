package service

import "errors"

var (
	ErrEmailAlreadyExists   = errors.New("EMAIL_ALREADY_EXISTS")
	ErrInvalidCredentials   = errors.New("INVALID_CREDENTIALS")
	ErrInvalidAdminSecret   = errors.New("INVALID_ADMIN_SECRET")
	ErrInvalidRefreshToken  = errors.New("INVALID_REFRESH_TOKEN")
	ErrAccessDenied         = errors.New("ACCESS_DENIED")
	ErrCreatorNotFound      = errors.New("CREATOR_NOT_FOUND")
	ErrVenueNotFound        = errors.New("VENUE_NOT_FOUND")
	ErrPhotoNotFound        = errors.New("PHOTO_NOT_FOUND")
	ErrProfileNotFound      = errors.New("PROFILE_NOT_FOUND")
	ErrProfileAlreadyExists = errors.New("PROFILE_ALREADY_EXISTS")
	ErrUserNotFound         = errors.New("USER_NOT_FOUND")
	ErrInvalidFileType             = errors.New("INVALID_FILE_TYPE")
	ErrFileTooLarge                = errors.New("FILE_TOO_LARGE")
	ErrInvalidImageType            = errors.New("INVALID_IMAGE_TYPE")
	ErrAlreadySubscribed           = errors.New("ALREADY_SUBSCRIBED")
	ErrInvalidUnsubscribeToken     = errors.New("INVALID_UNSUBSCRIBE_TOKEN")
	ErrAlreadyFavorited            = errors.New("ALREADY_FAVORITED")
	ErrFavoriteNotFound            = errors.New("FAVORITE_NOT_FOUND")
)
