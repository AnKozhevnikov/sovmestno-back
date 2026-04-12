package service

import (
	"application-service/internal/repository"
	"errors"
)

var (
	ErrCannotApplyToSelf             = errors.New("CANNOT_APPLY_TO_SELF")
	ErrDuplicatePendingApplication   = repository.ErrDuplicatePendingApplication
	ErrMirrorApplicationExists       = errors.New("MIRROR_APPLICATION_EXISTS")
	ErrApplicationNotFound           = errors.New("APPLICATION_NOT_FOUND")
	ErrAccessDenied                  = errors.New("ACCESS_DENIED")
	ErrApplicationAlreadyProcessed   = errors.New("APPLICATION_ALREADY_PROCESSED")
	ErrCollaborationNotFound         = errors.New("COLLABORATION_NOT_FOUND")
	ErrCollaborationAlreadyProcessed = errors.New("COLLABORATION_ALREADY_PROCESSED")
)
