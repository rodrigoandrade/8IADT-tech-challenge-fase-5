package repository

import "errors"

var (
	ErrLeadNotFound      = errors.New("lead not found")
	ErrLeadAlreadyExists = errors.New("lead already exists")

	ErrInvalidProperty       = errors.New("invalid property")
	ErrPropertyNotFound      = errors.New("property not found")
	ErrPropertyAlreadyExists = errors.New("property already exists")
)
